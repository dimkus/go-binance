package futures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Endpoints
const (
	baseWsMainUrl          = "wss://fstream.binance.com/ws"
	baseWsTestnetUrl       = "wss://stream.binancefuture.com/ws"
	baseCombinedMainURL    = "wss://fstream.binance.com/stream?streams="
	baseCombinedTestnetURL = "wss://stream.binancefuture.com/stream?streams="
)

var (
	// WebsocketTimeout is an interval for sending ping/pong messages if WebsocketKeepalive is enabled
	WebsocketTimeout = time.Second * 60
	// WebsocketKeepalive enables sending ping/pong messages to check the connection stability
	WebsocketKeepalive = false
	// UseTestnet switch all the WS streams from production to the testnet
	UseTestnet = false
	ProxyUrl   = ""
)

func getWsProxyUrl() *string {
	if ProxyUrl == "" {
		return nil
	}
	return &ProxyUrl
}

func SetWsProxyUrl(url string) {
	ProxyUrl = url
}

// getWsEndpoint return the base endpoint of the WS according the UseTestnet flag
func getWsEndpoint() string {
	if UseTestnet {
		return baseWsTestnetUrl
	}
	return baseWsMainUrl
}

// getCombinedEndpoint return the base endpoint of the combined stream according the UseTestnet flag
func getCombinedEndpoint() string {
	if UseTestnet {
		return baseCombinedTestnetURL
	}
	return baseCombinedMainURL
}

// WsAggTradeEventMessage define websocket aggTrde event message. fieldalignment
type WsAggTradeEventMessage struct {
	Data   *WsAggTradeEvent `json:"data"`
	Stream string           `json:"stream"`
}

// clear the fields of WsAggTradeEventMessage
func (e *WsAggTradeEventMessage) clear() {
	e.Stream = ""
	if e.Data != nil {
		e.Data.clear()
	}
}

// WsAggTradeEvent define websocket aggTrde event. fieldalignment
type WsAggTradeEvent struct {
	Symbol           string `json:"s"`
	Event            string `json:"e"`
	Price            string `json:"p"`
	Quantity         string `json:"q"`
	Time             int64  `json:"E"`
	AggregateTradeID int64  `json:"a"`
	FirstTradeID     int64  `json:"f"`
	LastTradeID      int64  `json:"l"`
	TradeTime        int64  `json:"T"`
	Maker            bool   `json:"m"`
}

// clear the fields of WsAggTradeEvent
func (e *WsAggTradeEvent) clear() {
	e.Event = ""
	e.Time = 0
	e.Symbol = ""
	e.AggregateTradeID = 0
	e.Price = ""
	e.Quantity = ""
	e.FirstTradeID = 0
	e.LastTradeID = 0
	e.TradeTime = 0
	e.Maker = false
}

// WsAggTradeHandler handle websocket that push trade information that is aggregated for a single taker order.
type WsAggTradeHandler func(event *WsAggTradeEvent)

// WsAggTradeServe serve websocket that push trade information that is aggregated for a single taker order.
func WsAggTradeServe(symbol string, handler WsAggTradeHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s@aggTrade", getWsEndpoint(), strings.ToLower(symbol))
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsAggTradeEvent)
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsCombinedAggTradeServe is similar to WsAggTradeServe, but it handles multiple symbols
func WsCombinedAggTradeServe(symbols []string, handler WsAggTradeHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := getCombinedEndpoint()
	for _, s := range symbols {
		endpoint += fmt.Sprintf("%s@aggTrade", strings.ToLower(s)) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}

		stream := j.Get("stream").MustString()
		data := j.Get("data").MustMap()

		symbol := strings.Split(stream, "@")[0]

		jsonData, _ := json.Marshal(data)

		event := new(WsAggTradeEvent)
		err = json.Unmarshal(jsonData, event)
		if err != nil {
			errHandler(err)
			return
		}
		event.Symbol = strings.ToUpper(symbol)

		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsCombinedAggTradeServeWithContext is similar to WsAggTradeServe, but it handles multiple symbols
func WsCombinedAggTradeServeWithContext(ctx context.Context, symbols []string, handler WsAggTradeHandler, errHandler ErrHandler) (err error) {
	endpoint := getCombinedEndpoint()
	for _, s := range symbols {
		endpoint += fmt.Sprintf("%s@aggTrade", strings.ToLower(s)) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}

		stream := j.Get("stream").MustString()
		data := j.Get("data").MustMap()

		symbol := strings.Split(stream, "@")[0]

		jsonData, _ := json.Marshal(data)

		event := new(WsAggTradeEvent)
		err = json.Unmarshal(jsonData, event)
		if err != nil {
			errHandler(err)
			return
		}
		event.Symbol = strings.ToUpper(symbol)

		handler(event)
	}
	return wsServeWithContext(ctx, cfg, wsHandler, errHandler)
}

var wsAggTradeEventMessageSyncPool = sync.Pool{
	New: func() any {
		return &WsAggTradeEventMessage{
			Data: new(WsAggTradeEvent),
		}
	},
}

// WsCombinedAggTradeServeWithPoolContext is similar to WsAggTradeServe, but it handles multiple symbols
// IMPORTANT: the WsAggTradeEvent pointer passed to WsAggTradeHandler is only valid during the callback execution.
// To preserve data, you must copy it. You should not store the pointer itself.
func WsCombinedAggTradeServeWithPoolContext(ctx context.Context, symbols []string, handler WsAggTradeHandler, errHandler ErrHandler) (err error) {
	endpoint := getCombinedEndpoint()
	for _, s := range symbols {
		endpoint += fmt.Sprintf("%s@aggTrade", strings.ToLower(s)) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		eventMessage := wsAggTradeEventMessageSyncPool.Get().(*WsAggTradeEventMessage)
		eventMessage.clear()

		defer wsAggTradeEventMessageSyncPool.Put(eventMessage)

		err = jsoniter.Unmarshal(message, eventMessage)
		if err != nil {
			errHandler(err)
			return
		}

		var symbol string
		if atIndex := strings.Index(eventMessage.Stream, "@"); atIndex != -1 {
			symbol = eventMessage.Stream[:atIndex]
		} else {
			symbol = eventMessage.Stream
		}

		eventMessage.Data.Symbol = strings.ToUpper(symbol)

		handler(eventMessage.Data)
	}
	return wsServeWithPoolContext(ctx, cfg, wsHandler, errHandler)
}

// WsMarkPriceEvent define websocket markPriceUpdate event.
type WsMarkPriceEvent struct {
	Event                string `json:"e"`
	Symbol               string `json:"s"`
	MarkPrice            string `json:"p"`
	IndexPrice           string `json:"i"`
	EstimatedSettlePrice string `json:"P"`
	FundingRate          string `json:"r"`
	Time                 int64  `json:"E"`
	NextFundingTime      int64  `json:"T"`
}

// WsMarkPriceHandler handle websocket that pushes price and funding rate for a single symbol.
type WsMarkPriceHandler func(event *WsMarkPriceEvent)

func wsMarkPriceServe(endpoint string, handler WsMarkPriceHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsMarkPriceEvent)
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsMarkPriceServe serve websocket that pushes price and funding rate for a single symbol.
func WsMarkPriceServe(symbol string, handler WsMarkPriceHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s@markPrice", getWsEndpoint(), strings.ToLower(symbol))
	return wsMarkPriceServe(endpoint, handler, errHandler)
}

// WsMarkPriceServeWithRate serve websocket that pushes price and funding rate for a single symbol and rate.
func WsMarkPriceServeWithRate(symbol string, rate time.Duration, handler WsMarkPriceHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	var rateStr string
	switch rate {
	case 3 * time.Second:
		rateStr = ""
	case 1 * time.Second:
		rateStr = "@1s"
	default:
		return nil, nil, errors.New("Invalid rate")
	}
	endpoint := fmt.Sprintf("%s/%s@markPrice%s", getWsEndpoint(), strings.ToLower(symbol), rateStr)
	return wsMarkPriceServe(endpoint, handler, errHandler)
}

func wsCombinedMarkPriceServe(endpoint string, handler WsMarkPriceHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}

		data := j.Get("data").MustMap()
		jsonData, _ := json.Marshal(data)

		event := new(WsMarkPriceEvent)
		err = json.Unmarshal(jsonData, event)
		if err != nil {
			errHandler(err)
			return
		}

		handler(event)
	}

	return wsServe(cfg, wsHandler, errHandler)
}

// WsCombinedMarkPriceServe is similar to WsMarkPriceServe, but it handles multiple symbols
func WsCombinedMarkPriceServe(symbols []string, handler WsMarkPriceHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := getCombinedEndpoint()
	for _, s := range symbols {
		endpoint += fmt.Sprintf("%s@markPrice", strings.ToLower(s)) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]

	return wsCombinedMarkPriceServe(endpoint, handler, errHandler)
}

// WsCombinedMarkPriceServeWithRate is similar to WsMarkPriceServeWithRate, but it for multiple symbols
func WsCombinedMarkPriceServeWithRate(symbolLevels map[string]time.Duration, handler WsMarkPriceHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := getCombinedEndpoint()
	for symbol, rate := range symbolLevels {
		var rateStr string
		switch rate {
		case 3 * time.Second:
			rateStr = ""
		case 1 * time.Second:
			rateStr = "@1s"
		default:
			return nil, nil, fmt.Errorf("invalid rate. Symbol %s (rate %d)", symbol, rate)
		}

		endpoint += fmt.Sprintf("%s@markPrice%s", strings.ToLower(symbol), rateStr) + "/"
	}

	endpoint = endpoint[:len(endpoint)-1]

	return wsCombinedMarkPriceServe(endpoint, handler, errHandler)
}

// WsAllMarkPriceEvent defines an array of websocket markPriceUpdate events.
type WsAllMarkPriceEvent []*WsMarkPriceEvent

// WsAllMarkPriceHandler handle websocket that pushes price and funding rate for all symbol.
type WsAllMarkPriceHandler func(event WsAllMarkPriceEvent)

func wsAllMarkPriceServe(endpoint string, handler WsAllMarkPriceHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		var event WsAllMarkPriceEvent
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsAllMarkPriceServe serve websocket that pushes price and funding rate for all symbol.
func WsAllMarkPriceServe(handler WsAllMarkPriceHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/!markPrice@arr", getWsEndpoint())
	return wsAllMarkPriceServe(endpoint, handler, errHandler)
}

// WsAllMarkPriceServeWithRate serve websocket that pushes price and funding rate for all symbol and rate.
func WsAllMarkPriceServeWithRate(rate time.Duration, handler WsAllMarkPriceHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	var rateStr string
	switch rate {
	case 3 * time.Second:
		rateStr = ""
	case 1 * time.Second:
		rateStr = "@1s"
	default:
		return nil, nil, errors.New("Invalid rate")
	}
	endpoint := fmt.Sprintf("%s/!markPrice@arr%s", getWsEndpoint(), rateStr)
	return wsAllMarkPriceServe(endpoint, handler, errHandler)
}

const (
	// initialWsAllMarkPriceEventPoolCapacity is the initial capacity of wsAllMarkPriceEventSyncPool
	initialWsAllMarkPriceEventPoolCapacity = 40
	// maxWsAllMarkPriceEventPoolCapacity is the maximum capacity of wsAllMarkPriceEventSyncPool
	// Don't put excessively large buffers back into the pool.
	maxWsAllMarkPriceEventPoolCapacity = 2000
)

// wsAllMarkPriceEventSyncPool is a sync.Pool for WsAllMarkPriceEvent
var wsAllMarkPriceEventSyncPool = sync.Pool{
	New: func() any {
		return make([]*WsMarkPriceEvent, 0, initialWsAllMarkPriceEventPoolCapacity)
	},
}

// WsAllMarkPriceServeWithPoolContext serve websocket that pushes price and funding rate for all symbol and rate
// use sync.Pool
// IMPORTANT: the WsAllMarkPriceEvent pointer passed to closure is only valid during the callback execution.
// To preserve data, you must copy it. You should not store the pointer itself.
func WsAllMarkPriceServeWithPoolContext(ctx context.Context, rate time.Duration, handler WsAllMarkPriceHandler, errHandler ErrHandler) error {
	var rateStr string
	switch rate {
	case 3 * time.Second:
		rateStr = ""
	case 1 * time.Second:
		rateStr = "@1s"
	default:
		return errors.New("Invalid rate")
	}

	endpoint := fmt.Sprintf("%s/!markPrice@arr%s", getWsEndpoint(), rateStr)

	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		events := wsAllMarkPriceEventSyncPool.Get().([]*WsMarkPriceEvent)

		// clear old events
		if len(events) != 0 {
			events = events[:0]
		}

		defer func() {
			// Don't put excessively large buffers back into the pool.
			if cap(events) <= maxWsAllMarkPriceEventPoolCapacity {
				for i := range events {
					events[i] = nil
				}
				wsAllMarkPriceEventSyncPool.Put(events)
			}
		}()

		err := jsoniter.Unmarshal(message, &events)
		if err != nil {
			errHandler(err)
			return
		}
		handler(events)
	}
	return wsServeWithPoolContext(ctx, cfg, wsHandler, errHandler)
}

// WsKlineEvent define websocket kline event
type WsKlineEvent struct {
	Event  string  `json:"e"`
	Symbol string  `json:"s"`
	Kline  WsKline `json:"k"`
	Time   int64   `json:"E"`
}

// WsKline define websocket kline
type WsKline struct {
	Volume               string `json:"v"`
	ActiveBuyQuoteVolume string `json:"Q"`
	Symbol               string `json:"s"`
	Interval             string `json:"i"`
	ActiveBuyVolume      string `json:"V"`
	QuoteVolume          string `json:"q"`
	Open                 string `json:"o"`
	Close                string `json:"c"`
	High                 string `json:"h"`
	Low                  string `json:"l"`
	FirstTradeID         int64  `json:"f"`
	TradeNum             int64  `json:"n"`
	LastTradeID          int64  `json:"L"`
	StartTime            int64  `json:"t"`
	EndTime              int64  `json:"T"`
	IsFinal              bool   `json:"x"`
}

// WsKlineHandler handle websocket kline event
type WsKlineHandler func(event *WsKlineEvent)

// WsKlineServe serve websocket kline handler with a symbol and interval like 15m, 30s
func WsKlineServe(symbol string, interval string, handler WsKlineHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s@kline_%s", getWsEndpoint(), strings.ToLower(symbol), interval)
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsKlineEvent)
		err := json.Unmarshal(message, event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsCombinedKlineServe is similar to WsKlineServe, but it handles multiple symbols with it interval
func WsCombinedKlineServe(symbolIntervalPair map[string]string, handler WsKlineHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := getCombinedEndpoint()
	for symbol, interval := range symbolIntervalPair {
		endpoint += fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}

		stream := j.Get("stream").MustString()
		data := j.Get("data").MustMap()

		symbol := strings.Split(stream, "@")[0]

		jsonData, _ := json.Marshal(data)

		event := new(WsKlineEvent)
		err = json.Unmarshal(jsonData, event)
		if err != nil {
			errHandler(err)
			return
		}
		event.Symbol = strings.ToUpper(symbol)

		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsCombinedKlineServeWithContext is similar to WsCombinedKlineServe, but it uses the given context.Context to control the lifetime of the websocket connection.
// It will serve websocket kline handler with multiple symbols with it interval.
// Each symbol will be served in a separate goroutine.
// The error returned is the first error encountered while establishing the websocket connection.
func WsCombinedKlineServeWithContext(ctx context.Context, symbolIntervalPair map[string]string, handler WsKlineHandler, errHandler ErrHandler) (err error) {
	endpoint := getCombinedEndpoint()
	for symbol, interval := range symbolIntervalPair {
		endpoint += fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}

		stream := j.Get("stream").MustString()
		data := j.Get("data").MustMap()

		symbol := strings.Split(stream, "@")[0]

		jsonData, _ := json.Marshal(data)

		event := new(WsKlineEvent)
		err = json.Unmarshal(jsonData, event)
		if err != nil {
			errHandler(err)
			return
		}
		event.Symbol = strings.ToUpper(symbol)

		handler(event)
	}
	return wsServeWithContext(ctx, cfg, wsHandler, errHandler)
}

// WsContinuousKlineEvent define websocket continuous kline event
type WsContinuousKlineEvent struct {
	Event        string            `json:"e"`
	PairSymbol   string            `json:"ps"`
	ContractType string            `json:"ct"`
	Kline        WsContinuousKline `json:"k"`
	Time         int64             `json:"E"`
}

// WsContinuousKline define websocket continuous kline
type WsContinuousKline struct {
	Volume               string `json:"v"`
	High                 string `json:"h"`
	Interval             string `json:"i"`
	ActiveBuyQuoteVolume string `json:"Q"`
	Low                  string `json:"l"`
	Open                 string `json:"o"`
	ActiveBuyVolume      string `json:"V"`
	QuoteVolume          string `json:"q"`
	Close                string `json:"c"`
	LastTradeID          int64  `json:"L"`
	StartTime            int64  `json:"t"`
	TradeNum             int64  `json:"n"`
	EndTime              int64  `json:"T"`
	FirstTradeID         int64  `json:"f"`
	IsFinal              bool   `json:"x"`
}

// WsContinuousKlineSubscribeArgs used with WsContinuousKlineServe or WsCombinedContinuousKlineServe
type WsContinuousKlineSubscribeArgs struct {
	Pair         string
	ContractType string
	Interval     string
}

// WsContinuousKlineHandler handle websocket continuous kline event
type WsContinuousKlineHandler func(event *WsContinuousKlineEvent)

// WsContinuousKlineServe serve websocket continuous kline handler with a pair and contractType and interval like 15m, 30s
func WsContinuousKlineServe(subscribeArgs *WsContinuousKlineSubscribeArgs, handler WsContinuousKlineHandler,
	errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s_%s@continuousKline_%s", getWsEndpoint(), strings.ToLower(subscribeArgs.Pair),
		strings.ToLower(subscribeArgs.ContractType), subscribeArgs.Interval)
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsContinuousKlineEvent)
		err := json.Unmarshal(message, event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsCombinedContinuousKlineServe is similar to WsContinuousKlineServe, but it handles multiple pairs of different contractType with its interval
func WsCombinedContinuousKlineServe(subscribeArgsList []*WsContinuousKlineSubscribeArgs,
	handler WsContinuousKlineHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := getCombinedEndpoint()
	for _, val := range subscribeArgsList {
		endpoint += fmt.Sprintf("%s_%s@continuousKline_%s", strings.ToLower(val.Pair),
			strings.ToLower(val.ContractType), val.Interval) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}

		data := j.Get("data").MustMap()

		jsonData, _ := json.Marshal(data)

		event := new(WsContinuousKlineEvent)
		err = json.Unmarshal(jsonData, event)
		if err != nil {
			errHandler(err)
			return
		}

		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsMiniMarketTickerEvent define websocket mini market ticker event.
type WsMiniMarketTickerEvent struct {
	Event       string `json:"e"`
	Symbol      string `json:"s"`
	ClosePrice  string `json:"c"`
	OpenPrice   string `json:"o"`
	HighPrice   string `json:"h"`
	LowPrice    string `json:"l"`
	Volume      string `json:"v"`
	QuoteVolume string `json:"q"`
	Time        int64  `json:"E"`
}

// WsMiniMarketTickerHandler handle websocket that pushes 24hr rolling window mini-ticker statistics for a single symbol.
type WsMiniMarketTickerHandler func(event *WsMiniMarketTickerEvent)

// WsMiniMarketTickerServe serve websocket that pushes 24hr rolling window mini-ticker statistics for a single symbol.
func WsMiniMarketTickerServe(symbol string, handler WsMiniMarketTickerHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s@miniTicker", getWsEndpoint(), strings.ToLower(symbol))
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsMiniMarketTickerEvent)
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsAllMiniMarketTickerEvent define an array of websocket mini market ticker events.
type WsAllMiniMarketTickerEvent []*WsMiniMarketTickerEvent

// WsAllMiniMarketTickerHandler handle websocket that pushes price and funding rate for all markets.
type WsAllMiniMarketTickerHandler func(event WsAllMiniMarketTickerEvent)

// WsAllMiniMarketTickerServe serve websocket that pushes price and funding rate for all markets.
func WsAllMiniMarketTickerServe(handler WsAllMiniMarketTickerHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/!miniTicker@arr", getWsEndpoint())
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		var event WsAllMiniMarketTickerEvent
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsMarketTickerEvent define websocket market ticker event.
type WsMarketTickerEvent struct {
	OpenPrice          string `json:"o"`
	LowPrice           string `json:"l"`
	Symbol             string `json:"s"`
	PriceChange        string `json:"p"`
	PriceChangePercent string `json:"P"`
	WeightedAvgPrice   string `json:"w"`
	ClosePrice         string `json:"c"`
	CloseQty           string `json:"Q"`
	QuoteVolume        string `json:"q"`
	Event              string `json:"e"`
	HighPrice          string `json:"h"`
	BaseVolume         string `json:"v"`
	Time               int64  `json:"E"`
	OpenTime           int64  `json:"O"`
	CloseTime          int64  `json:"C"`
	FirstID            int64  `json:"F"`
	LastID             int64  `json:"L"`
	TradeCount         int64  `json:"n"`
}

// WsMarketTickerHandler handle websocket that pushes 24hr rolling window mini-ticker statistics for a single symbol.
type WsMarketTickerHandler func(event *WsMarketTickerEvent)

// WsMarketTickerServe serve websocket that pushes 24hr rolling window mini-ticker statistics for a single symbol.
func WsMarketTickerServe(symbol string, handler WsMarketTickerHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s@ticker", getWsEndpoint(), strings.ToLower(symbol))
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsMarketTickerEvent)
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsAllMarketTickerEvent define an array of websocket mini ticker events.
type WsAllMarketTickerEvent []*WsMarketTickerEvent

// WsAllMarketTickerHandler handle websocket that pushes price and funding rate for all markets.
type WsAllMarketTickerHandler func(event WsAllMarketTickerEvent)

// WsAllMarketTickerServe serve websocket that pushes price and funding rate for all markets.
func WsAllMarketTickerServe(handler WsAllMarketTickerHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/!ticker@arr", getWsEndpoint())
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		var event WsAllMarketTickerEvent
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsBookTickerEvent define websocket best book ticker event. fieldalignment
type WsBookTickerEvent struct {
	Event           string `json:"e"`
	Symbol          string `json:"s"`
	BestBidPrice    string `json:"b"`
	BestBidQty      string `json:"B"`
	BestAskPrice    string `json:"a"`
	BestAskQty      string `json:"A"`
	UpdateID        int64  `json:"u"`
	Time            int64  `json:"E"`
	TransactionTime int64  `json:"T"`
}

// clear the fields of WsBookTickerEvent
func (e *WsBookTickerEvent) clear() {
	e.Event = ""
	e.Symbol = ""
	e.BestBidPrice = ""
	e.BestBidQty = ""
	e.BestAskPrice = ""
	e.BestAskQty = ""
	e.UpdateID = 0
	e.Time = 0
	e.TransactionTime = 0
}

// WsBookTickerHandler handle websocket that pushes updates to the best bid or ask price or quantity in real-time for a specified symbol.
type WsBookTickerHandler func(event *WsBookTickerEvent)

// WsBookTickerServe serve websocket that pushes updates to the best bid or ask price or quantity in real-time for a specified symbol.
func WsBookTickerServe(symbol string, handler WsBookTickerHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s@bookTicker", getWsEndpoint(), strings.ToLower(symbol))
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsBookTickerEvent)
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsAllBookTickerServe serve websocket that pushes updates to the best bid or ask price or quantity in real-time for all symbols.
func WsAllBookTickerServe(handler WsBookTickerHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/!bookTicker", getWsEndpoint())
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsBookTickerEvent)
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsCombinedBookTickerServe is similar to WsBookTickerServe, but it is for multiple symbols
func WsCombinedBookTickerServe(symbols []string, handler WsBookTickerHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/", getWsEndpoint())
	for _, s := range symbols {
		endpoint += fmt.Sprintf("%s@bookTicker", strings.ToLower(s)) + "/"
	}

	endpoint = endpoint[:len(endpoint)-1]
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsBookTickerEvent)
		err := json.Unmarshal(message, event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// wsBookTickerEventSyncPool is a sync.Pool for WsBookTickerEvent
var wsBookTickerEventSyncPool = sync.Pool{
	New: func() any {
		return &WsBookTickerEvent{}
	},
}

// WsCombinedBookTickerServeWithPoolContext uses the same "@bookTicker" endpoint as WsCombinedBookTickerServe, but uses a sync.Pool to reduce allocations
//
// IMPORTANT: the WsBookTickerEvent pointer passed to WsBookTickerHandler is only valid during the callback execution.
// To preserve data, you must copy it. You should not store the pointer itself.
func WsCombinedBookTickerServeWithPoolContext(ctx context.Context, symbols []string, handler WsBookTickerHandler, errHandler ErrHandler) (err error) {
	endpoint := fmt.Sprintf("%s/", getWsEndpoint())
	for _, s := range symbols {
		endpoint += fmt.Sprintf("%s@bookTicker", strings.ToLower(s)) + "/"
	}

	endpoint = endpoint[:len(endpoint)-1]
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := wsBookTickerEventSyncPool.Get().(*WsBookTickerEvent)
		event.clear()

		defer wsBookTickerEventSyncPool.Put(event)
		err = jsoniter.Unmarshal(message, event)
		if err != nil {
			errHandler(err)
			return
		}

		handler(event)
	}
	return wsServeWithPoolContext(ctx, cfg, wsHandler, errHandler)
}

// WsLiquidationOrderEvent define websocket liquidation order event.
type WsLiquidationOrderEvent struct {
	Event            string             `json:"e"`
	LiquidationOrder WsLiquidationOrder `json:"o"`
	Time             int64              `json:"E"`
}

// WsLiquidationOrder define websocket liquidation order.
type WsLiquidationOrder struct {
	Symbol               string          `json:"s"`
	Side                 SideType        `json:"S"`
	OrderType            OrderType       `json:"o"`
	TimeInForce          TimeInForceType `json:"f"`
	OrigQuantity         string          `json:"q"`
	Price                string          `json:"p"`
	AvgPrice             string          `json:"ap"`
	OrderStatus          OrderStatusType `json:"X"`
	LastFilledQty        string          `json:"l"`
	AccumulatedFilledQty string          `json:"z"`
	TradeTime            int64           `json:"T"`
}

// WsLiquidationOrderHandler handle websocket that pushes force liquidation order information for specific symbol.
type WsLiquidationOrderHandler func(event *WsLiquidationOrderEvent)

// WsLiquidationOrderServe serve websocket that pushes force liquidation order information for specific symbol.
func WsLiquidationOrderServe(symbol string, handler WsLiquidationOrderHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s@forceOrder", getWsEndpoint(), strings.ToLower(symbol))
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsLiquidationOrderEvent)
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsAllLiquidationOrderServe serve websocket that pushes force liquidation order information for all symbols.
func WsAllLiquidationOrderServe(handler WsLiquidationOrderHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/!forceOrder@arr", getWsEndpoint())
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsLiquidationOrderEvent)
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsDepthEvent define websocket depth book event
type WsDepthEvent struct {
	Event            string `json:"e"`
	Symbol           string `json:"s"`
	Bids             []Bid  `json:"b"`
	Asks             []Ask  `json:"a"`
	Time             int64  `json:"E"`
	TransactionTime  int64  `json:"T"`
	FirstUpdateID    int64  `json:"U"`
	LastUpdateID     int64  `json:"u"`
	PrevLastUpdateID int64  `json:"pu"`
}

// WsDepthHandler handle websocket depth event
type WsDepthHandler func(event *WsDepthEvent)

func wsPartialDepthServe(symbol string, levels int, rate *time.Duration, handler WsDepthHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	if levels != 5 && levels != 10 && levels != 20 {
		return nil, nil, errors.New("Invalid levels")
	}
	levelsStr := fmt.Sprintf("%d", levels)
	return wsDepthServe(symbol, levelsStr, rate, handler, errHandler)
}

// WsPartialDepthServe serve websocket partial depth handler.
func WsPartialDepthServe(symbol string, levels int, handler WsDepthHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	return wsPartialDepthServe(symbol, levels, nil, handler, errHandler)
}

// WsPartialDepthServeWithRate serve websocket partial depth handler with rate.
func WsPartialDepthServeWithRate(symbol string, levels int, rate time.Duration, handler WsDepthHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	return wsPartialDepthServe(symbol, levels, &rate, handler, errHandler)
}

// WsDiffDepthServe serve websocket diff. depth handler.
func WsDiffDepthServe(symbol string, handler WsDepthHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	return wsDepthServe(symbol, "", nil, handler, errHandler)
}

// WsCombinedDepthServe is similar to WsPartialDepthServe, but it for multiple symbols
func WsCombinedDepthServe(symbolLevels map[string]string, handler WsDepthHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := getCombinedEndpoint()
	for s, l := range symbolLevels {
		endpoint += fmt.Sprintf("%s@depth%s", strings.ToLower(s), l) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}
		event := new(WsDepthEvent)
		data := j.Get("data").MustMap()
		event.Event = data["e"].(string)
		event.Time, _ = data["E"].(json.Number).Int64()
		event.TransactionTime, _ = data["T"].(json.Number).Int64()
		event.Symbol = data["s"].(string)
		event.FirstUpdateID, _ = data["U"].(json.Number).Int64()
		event.LastUpdateID, _ = data["u"].(json.Number).Int64()
		event.PrevLastUpdateID, _ = data["pu"].(json.Number).Int64()
		bidsLen := len(data["b"].([]interface{}))
		event.Bids = make([]Bid, bidsLen)
		for i := 0; i < bidsLen; i++ {
			item := data["b"].([]interface{})[i].([]interface{})
			event.Bids[i] = Bid{
				Price:    item[0].(string),
				Quantity: item[1].(string),
			}
		}
		asksLen := len(data["a"].([]interface{}))
		event.Asks = make([]Ask, asksLen)
		for i := 0; i < asksLen; i++ {
			item := data["a"].([]interface{})[i].([]interface{})
			event.Asks[i] = Ask{
				Price:    item[0].(string),
				Quantity: item[1].(string),
			}
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsCombinedDiffDepthServe is similar to WsDiffDepthServe, but it for multiple symbols
func WsCombinedDiffDepthServe(symbols []string, handler WsDepthHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := getCombinedEndpoint()
	for _, s := range symbols {
		endpoint += fmt.Sprintf("%s@depth", strings.ToLower(s)) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}
		event := new(WsDepthEvent)
		data := j.Get("data").MustMap()
		event.Event = data["e"].(string)
		event.Time, _ = data["E"].(json.Number).Int64()
		event.TransactionTime, _ = data["T"].(json.Number).Int64()
		event.Symbol = data["s"].(string)
		event.FirstUpdateID, _ = data["U"].(json.Number).Int64()
		event.LastUpdateID, _ = data["u"].(json.Number).Int64()
		event.PrevLastUpdateID, _ = data["pu"].(json.Number).Int64()
		bidsLen := len(data["b"].([]interface{}))
		event.Bids = make([]Bid, bidsLen)
		for i := 0; i < bidsLen; i++ {
			item := data["b"].([]interface{})[i].([]interface{})
			event.Bids[i] = Bid{
				Price:    item[0].(string),
				Quantity: item[1].(string),
			}
		}
		asksLen := len(data["a"].([]interface{}))
		event.Asks = make([]Ask, asksLen)
		for i := 0; i < asksLen; i++ {
			item := data["a"].([]interface{})[i].([]interface{})
			event.Asks[i] = Ask{
				Price:    item[0].(string),
				Quantity: item[1].(string),
			}
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsDiffDepthServeWithRate serve websocket diff. depth handler with rate.
func WsDiffDepthServeWithRate(symbol string, rate time.Duration, handler WsDepthHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	return wsDepthServe(symbol, "", &rate, handler, errHandler)
}

func wsDepthServe(symbol string, levels string, rate *time.Duration, handler WsDepthHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	var rateStr string
	if rate != nil {
		switch *rate {
		case 250 * time.Millisecond:
			rateStr = ""
		case 500 * time.Millisecond:
			rateStr = "@500ms"
		case 100 * time.Millisecond:
			rateStr = "@100ms"
		default:
			return nil, nil, errors.New("Invalid rate")
		}
	}
	endpoint := fmt.Sprintf("%s/%s@depth%s%s", getWsEndpoint(), strings.ToLower(symbol), levels, rateStr)
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}
		event := new(WsDepthEvent)
		event.Event = j.Get("e").MustString()
		event.Time = j.Get("E").MustInt64()
		event.TransactionTime = j.Get("T").MustInt64()
		event.Symbol = j.Get("s").MustString()
		event.FirstUpdateID = j.Get("U").MustInt64()
		event.LastUpdateID = j.Get("u").MustInt64()
		event.PrevLastUpdateID = j.Get("pu").MustInt64()
		bidsLen := len(j.Get("b").MustArray())
		event.Bids = make([]Bid, bidsLen)
		for i := 0; i < bidsLen; i++ {
			item := j.Get("b").GetIndex(i)
			event.Bids[i] = Bid{
				Price:    item.GetIndex(0).MustString(),
				Quantity: item.GetIndex(1).MustString(),
			}
		}
		asksLen := len(j.Get("a").MustArray())
		event.Asks = make([]Ask, asksLen)
		for i := 0; i < asksLen; i++ {
			item := j.Get("a").GetIndex(i)
			event.Asks[i] = Ask{
				Price:    item.GetIndex(0).MustString(),
				Quantity: item.GetIndex(1).MustString(),
			}
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsBLVTInfoEvent define websocket BLVT info event
type WsBLVTInfoEvent struct {
	Event          string         `json:"e"`
	Symbol         string         `json:"s"`
	Baskets        []WsBLVTBasket `json:"b"`
	Time           int64          `json:"E"`
	Issued         float64        `json:"m"`
	Nav            float64        `json:"n"`
	Leverage       float64        `json:"l"`
	TargetLeverage int64          `json:"t"`
	FundingRate    float64        `json:"f"`
}

// WsBLVTBasket define websocket BLVT basket
type WsBLVTBasket struct {
	Symbol   string `json:"s"`
	Position int64  `json:"n"`
}

// WsBLVTInfoHandler handle websocket BLVT event
type WsBLVTInfoHandler func(event *WsBLVTInfoEvent)

// WsBLVTInfoServe serve BLVT info stream
func WsBLVTInfoServe(name string, handler WsBLVTInfoHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s@tokenNav", getWsEndpoint(), strings.ToUpper(name))
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsBLVTInfoEvent)
		err := json.Unmarshal(message, &event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsBLVTKlineEvent define BLVT kline event
type WsBLVTKlineEvent struct {
	Event  string      `json:"e"`
	Symbol string      `json:"s"`
	Kline  WsBLVTKline `json:"k"`
	Time   int64       `json:"E"`
}

// WsBLVTKline BLVT kline
type WsBLVTKline struct {
	Symbol          string `json:"s"`
	Interval        string `json:"i"`
	OpenPrice       string `json:"o"`
	ClosePrice      string `json:"c"`
	HighPrice       string `json:"h"`
	LowPrice        string `json:"l"`
	Leverage        string `json:"v"`
	StartTime       int64  `json:"t"`
	CloseTime       int64  `json:"T"`
	FirstUpdateTime int64  `json:"f"`
	LastUpdateTime  int64  `json:"L"`
	Count           int64  `json:"n"`
}

// WsBLVTKlineHandler BLVT kline handler
type WsBLVTKlineHandler func(event *WsBLVTKlineEvent)

// WsBLVTKlineServe serve BLVT kline stream
func WsBLVTKlineServe(name string, interval string, handler WsBLVTKlineHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s@nav_Kline_%s", getWsEndpoint(), strings.ToUpper(name), interval)
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsBLVTKlineEvent)
		err := json.Unmarshal(message, event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsCompositeIndexEvent websocket composite index event
type WsCompositeIndexEvent struct {
	Event       string          `json:"e"`
	Symbol      string          `json:"s"`
	Price       string          `json:"p"`
	Composition []WsComposition `json:"c"`
	Time        int64           `json:"E"`
}

// WsComposition websocket composite index event composition
type WsComposition struct {
	BaseAsset    string `json:"b"`
	WeightQty    string `json:"w"`
	WeighPercent string `json:"W"`
}

// WsCompositeIndexHandler websocket composite index handler
type WsCompositeIndexHandler func(event *WsCompositeIndexEvent)

// WsCompositiveIndexServe serve composite index information for index symbols
func WsCompositiveIndexServe(symbol string, handler WsCompositeIndexHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s@compositeIndex", getWsEndpoint(), strings.ToLower(symbol))
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {
		event := new(WsCompositeIndexEvent)
		err := json.Unmarshal(message, event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsUserDataEvent define user data event
type WsUserDataEvent struct {
	Event               UserDataEventType     `json:"e"`
	CrossWalletBalance  string                `json:"cw"`
	AccountUpdate       WsAccountUpdate       `json:"a"`
	MarginCallPositions []WsPosition          `json:"p"`
	AccountConfigUpdate WsAccountConfigUpdate `json:"ac"`
	OrderTradeUpdate    WsOrderTradeUpdate    `json:"o"`
	Time                int64                 `json:"E"`
	TransactionTime     int64                 `json:"T"`
}

func (e *WsUserDataEvent) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Time                interface{}           `json:"E"`
		Event               UserDataEventType     `json:"e"`
		CrossWalletBalance  string                `json:"cw"`
		AccountUpdate       WsAccountUpdate       `json:"a"`
		MarginCallPositions []WsPosition          `json:"p"`
		AccountConfigUpdate WsAccountConfigUpdate `json:"ac"`
		OrderTradeUpdate    WsOrderTradeUpdate    `json:"o"`
		TransactionTime     int64                 `json:"T"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	e.Event = tmp.Event
	switch v := tmp.Time.(type) {
	case float64:
		e.Time = int64(v)
	case string:
		parsedTime, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		e.Time = parsedTime
	default:
		return fmt.Errorf("unexpected type for E: %T", tmp.Time)
	}
	e.CrossWalletBalance = tmp.CrossWalletBalance
	e.MarginCallPositions = tmp.MarginCallPositions
	e.TransactionTime = tmp.TransactionTime
	e.AccountUpdate = tmp.AccountUpdate
	e.OrderTradeUpdate = tmp.OrderTradeUpdate
	e.AccountConfigUpdate = tmp.AccountConfigUpdate
	return nil
}

// WsAccountUpdate define account update
type WsAccountUpdate struct {
	Reason    UserDataEventReasonType `json:"m"`
	Balances  []WsBalance             `json:"B"`
	Positions []WsPosition            `json:"P"`
}

// WsBalance define balance
type WsBalance struct {
	Asset              string `json:"a"`
	Balance            string `json:"wb"`
	CrossWalletBalance string `json:"cw"`
	ChangeBalance      string `json:"bc"`
}

// WsPosition define position
type WsPosition struct {
	Symbol                    string           `json:"s"`
	Side                      PositionSideType `json:"ps"`
	Amount                    string           `json:"pa"`
	MarginType                MarginType       `json:"mt"`
	IsolatedWallet            string           `json:"iw"`
	EntryPrice                string           `json:"ep"`
	MarkPrice                 string           `json:"mp"`
	UnrealizedPnL             string           `json:"up"`
	AccumulatedRealized       string           `json:"cr"`
	MaintenanceMarginRequired string           `json:"mm"`
}

// WsOrderTradeUpdate define order trade update
type WsOrderTradeUpdate struct {
	Commission           string             `json:"n"`
	PositionSide         PositionSideType   `json:"ps"`
	Side                 SideType           `json:"S"`
	Type                 OrderType          `json:"o"`
	TimeInForce          TimeInForceType    `json:"f"`
	OriginalQty          string             `json:"q"`
	OriginalPrice        string             `json:"p"`
	AveragePrice         string             `json:"ap"`
	StopPrice            string             `json:"sp"`
	ExecutionType        OrderExecutionType `json:"x"`
	PriceMode            string             `json:"pm"`
	STP                  string             `json:"V"`
	LastFilledQty        string             `json:"l"`
	AccumulatedFilledQty string             `json:"z"`
	LastFilledPrice      string             `json:"L"`
	CommissionAsset      string             `json:"N"`
	ClientOrderID        string             `json:"c"`
	RealizedPnL          string             `json:"rp"`
	Status               OrderStatusType    `json:"X"`
	BidsNotional         string             `json:"b"`
	AsksNotional         string             `json:"a"`
	CallbackRate         string             `json:"cr"`
	ActivationPrice      string             `json:"AP"`
	WorkingType          WorkingType        `json:"wt"`
	OriginalType         OrderType          `json:"ot"`
	Symbol               string             `json:"s"`
	TradeTime            int64              `json:"T"`
	ID                   int64              `json:"i"`
	TradeID              int64              `json:"t"`
	GTD                  int64              `json:"gtd"`
	IsClosingPosition    bool               `json:"cp"`
	IsReduceOnly         bool               `json:"R"`
	IsMaker              bool               `json:"m"`
	PriceProtect         bool               `json:"pP"`
}

// WsAccountConfigUpdate define account config update
type WsAccountConfigUpdate struct {
	Symbol   string `json:"s"`
	Leverage int64  `json:"l"`
}

// WsUserDataHandler handle WsUserDataEvent
type WsUserDataHandler func(event *WsUserDataEvent)

// WsUserDataServe serve user data handler with listen key
func WsUserDataServe(listenKey string, handler WsUserDataHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s", getWsEndpoint(), listenKey)
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {

		/*
				FIXME: skip TRADE_LITE event
				https://www.binance.com/en/support/announcement/introducing-trade-lite-for-usd%E2%93%A2-m-futures-websocket-2024-09-03-f2809259702b46f7abdc9c97b977908f

				{"e":"ORDER_TRADE_UPDATE","T":1725462901305,"E":1725462901305,"o":{"s":"ETCUSDT","c":"x-hGBqJOcM1420269833_1945678308","S":"BUY","o":"LIMIT","f":"GTC","q":"350.10","p":"18.024","ap":"0","sp":"0","x":"NEW","X":"NEW","i":23178779928,"l":"0","z":"0","L":"0","n":"0","N":"USDT","T":1725462901305,"t":0,"b":"6310.20240","a":"0","m":false,"R":false,"wt":"CONTRACT_PRICE","ot":"LIMIT","ps":"BOTH","cp":false,"rp":"0","pP":false,"si":0,"ss":0,"V":"NONE","pm":"NONE","gtd":0}}
				{"e":"TRADE_LITE","E":1725462902696,"T":1725462902695,"s":"ETCUSDT","q":"350.10","p":"0.000","m":false,"c":"x-hGBqJOcM1420274314_2099272109","S":"BUY","L":"18.030","l":"52.74","t":937937664,"i":23178780330}

				{"e":"ORDER_TRADE_UPDATE","T":1725462901591,"E":1725462901591,"o":{"s":"ETCUSDT","c":"x-hGBqJOcM1420275256_949569752","S":"BUY","o":"LIMIT","f":"GTC","q":"271.79","p":"18.024","ap":"0","sp":"0","x":"NEW","X":"NEW","i":23178780068,"l":"0","z":"0","L":"0","n":"0","N":"USDT","T":1725462901591,"t":0,"b":"11208.94536","a":"0","m":false,"R":false,"wt":"CONTRACT_PRICE","ot":"LIMIT","ps":"BOTH","cp":false,"rp":"0","pP":false,"si":0,"ss":0,"V":"NONE","pm":"NONE","gtd":0}}
				{"e":"TRADE_LITE","E":1725462972436,"T":1725462972436,"s":"ETCUSDT","q":"271.79","p":"18.024","m":true,"c":"x-hGBqJOcM1420275256_949569752","S":"BUY","L":"18.024","l":"271.79","t":937937846,"i":23178780068}

			The "TRADE_LITE" event is sent by the FSTREAM-MM that was subscribed by the listenkey of Futures Market Makers (MM).
			If you want to get the price of your own trades, you can use the GET /fapi/v1/userTrades
			endpoint: https://binance-docs.github.io/apidocs/futures/en/#account-trade-list-user_data
				-admin
		*/
		if strings.Index(string(message), "\"TRADE_LITE\"") != -1 {
			return
		}

		event := new(WsUserDataEvent)
		err := json.Unmarshal(message, event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServe(cfg, wsHandler, errHandler)
}

// WsUserDataServeWithContext serve user data handler with listen key
func WsUserDataServeWithContext(ctx context.Context, listenKey string, handler WsUserDataHandler, errHandler ErrHandler) (err error) {
	endpoint := fmt.Sprintf("%s/%s", getWsEndpoint(), listenKey)
	cfg := newWsConfig(endpoint)
	wsHandler := func(message []byte) {

		/*
				FIXME: skip TRADE_LITE event
				https://www.binance.com/en/support/announcement/introducing-trade-lite-for-usd%E2%93%A2-m-futures-websocket-2024-09-03-f2809259702b46f7abdc9c97b977908f

				{"e":"ORDER_TRADE_UPDATE","T":1725462901305,"E":1725462901305,"o":{"s":"ETCUSDT","c":"x-hGBqJOcM1420269833_1945678308","S":"BUY","o":"LIMIT","f":"GTC","q":"350.10","p":"18.024","ap":"0","sp":"0","x":"NEW","X":"NEW","i":23178779928,"l":"0","z":"0","L":"0","n":"0","N":"USDT","T":1725462901305,"t":0,"b":"6310.20240","a":"0","m":false,"R":false,"wt":"CONTRACT_PRICE","ot":"LIMIT","ps":"BOTH","cp":false,"rp":"0","pP":false,"si":0,"ss":0,"V":"NONE","pm":"NONE","gtd":0}}
				{"e":"TRADE_LITE","E":1725462902696,"T":1725462902695,"s":"ETCUSDT","q":"350.10","p":"0.000","m":false,"c":"x-hGBqJOcM1420274314_2099272109","S":"BUY","L":"18.030","l":"52.74","t":937937664,"i":23178780330}

				{"e":"ORDER_TRADE_UPDATE","T":1725462901591,"E":1725462901591,"o":{"s":"ETCUSDT","c":"x-hGBqJOcM1420275256_949569752","S":"BUY","o":"LIMIT","f":"GTC","q":"271.79","p":"18.024","ap":"0","sp":"0","x":"NEW","X":"NEW","i":23178780068,"l":"0","z":"0","L":"0","n":"0","N":"USDT","T":1725462901591,"t":0,"b":"11208.94536","a":"0","m":false,"R":false,"wt":"CONTRACT_PRICE","ot":"LIMIT","ps":"BOTH","cp":false,"rp":"0","pP":false,"si":0,"ss":0,"V":"NONE","pm":"NONE","gtd":0}}
				{"e":"TRADE_LITE","E":1725462972436,"T":1725462972436,"s":"ETCUSDT","q":"271.79","p":"18.024","m":true,"c":"x-hGBqJOcM1420275256_949569752","S":"BUY","L":"18.024","l":"271.79","t":937937846,"i":23178780068}

			The "TRADE_LITE" event is sent by the FSTREAM-MM that was subscribed by the listenkey of Futures Market Makers (MM).
			If you want to get the price of your own trades, you can use the GET /fapi/v1/userTrades
			endpoint: https://binance-docs.github.io/apidocs/futures/en/#account-trade-list-user_data
				-admin
		*/
		if strings.Index(string(message), "\"TRADE_LITE\"") != -1 {
			return
		}

		event := new(WsUserDataEvent)
		err := json.Unmarshal(message, event)
		if err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return wsServeWithContext(ctx, cfg, wsHandler, errHandler)
}

func WsUserDataServeOld(listenKey string, handler WsHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	endpoint := fmt.Sprintf("%s/%s", getWsEndpoint(), listenKey)
	cfg := newWsConfig(endpoint)
	return wsServe(cfg, handler, errHandler)
}
