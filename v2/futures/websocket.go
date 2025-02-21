package futures

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// WsHandler handle raw websocket message
type WsHandler func(message []byte)

// ErrHandler handles errors
type ErrHandler func(err error)

// WsConfig webservice configuration
type WsConfig struct {
	Endpoint string
	Proxy    *string
}

func newWsConfig(endpoint string) *WsConfig {
	return &WsConfig{
		Endpoint: endpoint,
		Proxy:    getWsProxyUrl(),
	}
}

var wsServe = func(cfg *WsConfig, handler WsHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	proxy := http.ProxyFromEnvironment
	if cfg.Proxy != nil {
		u, err := url.Parse(*cfg.Proxy)
		if err != nil {
			return nil, nil, err
		}
		proxy = http.ProxyURL(u)
	}
	Dialer := websocket.Dialer{
		Proxy:             proxy,
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: false,
	}

	c, _, err := Dialer.Dial(cfg.Endpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	c.SetReadLimit(655350)
	doneC = make(chan struct{})
	stopC = make(chan struct{})
	go func() {
		// This function will exit either on error from
		// websocket.Conn.ReadMessage or when the stopC channel is
		// closed by the client.
		defer close(doneC)
		if WebsocketKeepalive {
			keepAlive(c, WebsocketTimeout)
		}
		// Wait for the stopC channel to be closed.  We do that in a
		// separate goroutine because ReadMessage is a blocking
		// operation.
		silent := false
		go func() {
			select {
			case <-stopC:
				silent = true
			case <-doneC:
			}
			c.Close()
		}()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if !silent {
					errHandler(err)
				}
				return
			}
			handler(message)
		}
	}()
	return
}

func keepAlive(c *websocket.Conn, timeout time.Duration) {
	ticker := time.NewTicker(timeout)

	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		defer ticker.Stop()
		for {
			deadline := time.Now().Add(10 * time.Second)
			err := c.WriteControl(websocket.PingMessage, []byte{}, deadline)
			if err != nil {
				return
			}
			<-ticker.C
			if time.Since(lastResponse) > timeout {
				c.Close()
				return
			}
		}
	}()
}

func wsServeWithContext(ctx context.Context, cfg *WsConfig, handler WsHandler, errHandler ErrHandler) (err error) {
	proxy := http.ProxyFromEnvironment
	if cfg.Proxy != nil {
		u, err := url.Parse(*cfg.Proxy)
		if err != nil {
			return err
		}
		proxy = http.ProxyURL(u)
	}
	dialer := websocket.Dialer{
		Proxy:             proxy,
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: false,
	}

	ctxToRun, ctxToRunCancel := context.WithCancel(ctx)

	conn, _, err := dialer.DialContext(ctxToRun, cfg.Endpoint, nil)
	if err != nil {
		ctxToRunCancel()
		return err
	}

	conn.SetReadLimit(655350)

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})

	go func(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) {
		defer cancel()

		go func(ctx context.Context, conn *websocket.Conn) {
			pingTicker := time.NewTicker(WebsocketTimeout)
			defer pingTicker.Stop()

			for {
				select {
				case <-ctx.Done():
					conn.Close()
					return
				case <-pingTicker.C:
					if !WebsocketKeepalive {
						continue
					}

					lastResponse := time.Now()
					conn.SetPongHandler(func(msg string) error {
						lastResponse = time.Now()
						return nil
					})

					deadline := time.Now().Add(10 * time.Second)
					pingErr := conn.WriteControl(websocket.PingMessage, []byte{}, deadline)
					if pingErr != nil {
						conn.Close()
						return
					}

					if time.Since(lastResponse) > WebsocketTimeout {
						conn.Close()
						return
					}
				}
			}
		}(ctx, conn)

		for {
			select {
			case <-ctx.Done():
				errHandler(ctx.Err())
				return
			default:
				_, message, readMessageErr := conn.ReadMessage()
				if readMessageErr != nil {
					errHandler(readMessageErr)
					return
				}
				handler(message)
			}
		}

	}(ctxToRun, ctxToRunCancel, conn)

	return nil
}
