package futures

import (
	"context"
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"sync"
)

// GetBalanceService get account balance
type GetBalanceService struct {
	c *Client
}

// Do send request
func (s *GetBalanceService) Do(ctx context.Context, opts ...RequestOption) (res []*Balance, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v2/balance",
		secType:  secTypeSigned,
	}
	data, _, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return []*Balance{}, err
	}
	res = make([]*Balance, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return []*Balance{}, err
	}
	return res, nil
}

// Balance define user balance of your account
type Balance struct {
	AccountAlias       string `json:"accountAlias"`
	Asset              string `json:"asset"`
	Balance            string `json:"balance"`
	CrossWalletBalance string `json:"crossWalletBalance"`
	CrossUnPnl         string `json:"crossUnPnl"`
	AvailableBalance   string `json:"availableBalance"`
	MaxWithdrawAmount  string `json:"maxWithdrawAmount"`
}

// GetAccountService get account info
type GetAccountService struct {
	c *Client
}

// Do send request
func (s *GetAccountService) Do(ctx context.Context, opts ...RequestOption) (res *Account, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v2/account",
		secType:  secTypeSigned,
	}
	data, _, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = new(Account)
	err = json.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// accountPool is a pool for account
var accountPool = &sync.Pool{
	New: func() interface{} {
		return &Account{
			Positions: make([]*AccountPosition, 0, 10),
			Assets:    make([]*AccountAsset, 0, 10),
		}
	},
}

// DoWithPool sends a request using a shared pool of Account objects
//
// Important: The *Account instance provided to the closure is borrowed from an object pool
// and is only valid during the closure execution. Do NOT retain references to the account
// or any of its nested data (including Assets and Positions slices) after the closure returns,
// as the object will be reclaimed by the pool and reused for subsequent requests.
//
// If an error occurs during the request, the account parameter will be nil and the error
// will be provided through the err parameter.
func (s *GetAccountService) DoWithPool(ctx context.Context, fn func(account *Account), opts ...RequestOption) error {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v2/account",
		secType:  secTypeSigned,
	}

	err := s.c.callApiWithPool(ctx, func(data []byte, _ *http.Header) error {
		account := accountPool.Get().(*Account)
		account.clear()
		defer accountPool.Put(account)

		errUnmarshal := jsoniter.Unmarshal(data, account)
		if errUnmarshal != nil {
			return errUnmarshal
		}
		fn(account)

		return nil
	}, r, opts...)

	if err != nil {
		return err
	}

	return nil
}

// Account define account info, fieldalignment
type Account struct {
	TotalInitialMargin          string             `json:"totalInitialMargin"`
	TotalMaintMargin            string             `json:"totalMaintMargin"`
	TotalWalletBalance          string             `json:"totalWalletBalance"`
	TotalUnrealizedProfit       string             `json:"totalUnrealizedProfit"`
	TotalMarginBalance          string             `json:"totalMarginBalance"`
	TotalPositionInitialMargin  string             `json:"totalPositionInitialMargin"`
	TotalOpenOrderInitialMargin string             `json:"totalOpenOrderInitialMargin"`
	TotalCrossWalletBalance     string             `json:"totalCrossWalletBalance"`
	TotalCrossUnPnl             string             `json:"totalCrossUnPnl"`
	AvailableBalance            string             `json:"availableBalance"`
	MaxWithdrawAmount           string             `json:"maxWithdrawAmount"`
	Assets                      []*AccountAsset    `json:"assets"`
	Positions                   []*AccountPosition `json:"positions"`
	UpdateTime                  int64              `json:"updateTime"`
	FeeTier                     int                `json:"feeTier"`
	CanTrade                    bool               `json:"canTrade"`
	CanDeposit                  bool               `json:"canDeposit"`
	CanWithdraw                 bool               `json:"canWithdraw"`
	MultiAssetsMargin           bool               `json:"multiAssetsMargin"`
}

// clear the fields of Account
func (a *Account) clear() {
	const (
		maxAssetsCapacity    = 500
		maxPositionsCapacity = 60
		defaultCapacity      = 10
	)

	a.FeeTier = 0
	a.CanTrade = false
	a.CanDeposit = false
	a.CanWithdraw = false
	a.UpdateTime = 0
	a.MultiAssetsMargin = false
	a.TotalInitialMargin = ""
	a.TotalMaintMargin = ""
	a.TotalWalletBalance = ""
	a.TotalUnrealizedProfit = ""
	a.TotalMarginBalance = ""
	a.TotalPositionInitialMargin = ""
	a.TotalOpenOrderInitialMargin = ""
	a.TotalCrossWalletBalance = ""
	a.TotalCrossUnPnl = ""
	a.AvailableBalance = ""
	a.MaxWithdrawAmount = ""

	if a.Assets != nil {
		if cap(a.Assets) > maxAssetsCapacity {
			a.Assets = make([]*AccountAsset, 0, defaultCapacity)
		} else {
			for i := range a.Assets {
				a.Assets[i] = nil
			}
			a.Assets = a.Assets[:0]
		}
	}

	if a.Positions != nil {
		if cap(a.Positions) > maxPositionsCapacity {
			a.Positions = make([]*AccountPosition, 0, defaultCapacity)
		} else {
			for i := range a.Positions {
				a.Positions[i] = nil
			}
			a.Positions = a.Positions[:0]
		}
	}
}

// AccountAsset define account asset
type AccountAsset struct {
	Asset                  string `json:"asset"`
	InitialMargin          string `json:"initialMargin"`
	MaintMargin            string `json:"maintMargin"`
	MarginBalance          string `json:"marginBalance"`
	MaxWithdrawAmount      string `json:"maxWithdrawAmount"`
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"`
	PositionInitialMargin  string `json:"positionInitialMargin"`
	UnrealizedProfit       string `json:"unrealizedProfit"`
	WalletBalance          string `json:"walletBalance"`
	CrossWalletBalance     string `json:"crossWalletBalance"`
	CrossUnPnl             string `json:"crossUnPnl"`
	AvailableBalance       string `json:"availableBalance"`
	MarginAvailable        bool   `json:"marginAvailable"`
	UpdateTime             int64  `json:"updateTime"`
}

// AccountPosition define account position, fieldalignment
type AccountPosition struct {
	Leverage               string           `json:"leverage"`
	InitialMargin          string           `json:"initialMargin"`
	MaintMargin            string           `json:"maintMargin"`
	OpenOrderInitialMargin string           `json:"openOrderInitialMargin"`
	PositionInitialMargin  string           `json:"positionInitialMargin"`
	Symbol                 string           `json:"symbol"`
	UnrealizedProfit       string           `json:"unrealizedProfit"`
	EntryPrice             string           `json:"entryPrice"`
	MaxNotional            string           `json:"maxNotional"`
	PositionAmt            string           `json:"positionAmt"`
	Notional               string           `json:"notional"`
	BidNotional            string           `json:"bidNotional"`
	AskNotional            string           `json:"askNotional"`
	PositionSide           PositionSideType `json:"positionSide"`
	UpdateTime             int64            `json:"updateTime"`
	Isolated               bool             `json:"isolated"`
}
