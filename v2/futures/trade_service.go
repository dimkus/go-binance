package futures

import (
	"context"
	"encoding/json"
	"net/http"
)

// HistoricalTradesService trades
type HistoricalTradesService struct {
	c      *Client
	limit  *int
	fromID *int64
	symbol string
}

// Symbol set symbol
func (s *HistoricalTradesService) Symbol(symbol string) *HistoricalTradesService {
	s.symbol = symbol
	return s
}

// Limit set limit
func (s *HistoricalTradesService) Limit(limit int) *HistoricalTradesService {
	s.limit = &limit
	return s
}

// FromID set fromID
func (s *HistoricalTradesService) FromID(fromID int64) *HistoricalTradesService {
	s.fromID = &fromID
	return s
}

// Do send request
func (s *HistoricalTradesService) Do(ctx context.Context, opts ...RequestOption) (res []*Trade, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/historicalTrades",
		secType:  secTypeAPIKey,
	}
	r.setParam("symbol", s.symbol)
	if s.limit != nil {
		r.setParam("limit", *s.limit)
	}
	if s.fromID != nil {
		r.setParam("fromId", *s.fromID)
	}

	data, _, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return
	}
	res = make([]*Trade, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return
	}
	return
}

// Trade define trade info
type Trade struct {
	Price         string `json:"price"`
	Quantity      string `json:"qty"`
	QuoteQuantity string `json:"quoteQty"`
	ID            int64  `json:"id"`
	Time          int64  `json:"time"`
	IsBuyerMaker  bool   `json:"isBuyerMaker"`
}

// TradeV3 define v3 trade info
type TradeV3 struct {
	Symbol          string `json:"symbol"`
	Price           string `json:"price"`
	Quantity        string `json:"qty"`
	QuoteQuantity   string `json:"quoteQty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	ID              int64  `json:"id"`
	OrderID         int64  `json:"orderId"`
	Time            int64  `json:"time"`
	IsBuyer         bool   `json:"isBuyer"`
	IsMaker         bool   `json:"isMaker"`
	IsBestMatch     bool   `json:"isBestMatch"`
}

// AggTradesService list aggregate trades
type AggTradesService struct {
	c         *Client
	fromID    *int64
	startTime *int64
	endTime   *int64
	limit     *int
	symbol    string
}

// Symbol set symbol
func (s *AggTradesService) Symbol(symbol string) *AggTradesService {
	s.symbol = symbol
	return s
}

// FromID set fromID
func (s *AggTradesService) FromID(fromID int64) *AggTradesService {
	s.fromID = &fromID
	return s
}

// StartTime set startTime
func (s *AggTradesService) StartTime(startTime int64) *AggTradesService {
	s.startTime = &startTime
	return s
}

// EndTime set endTime
func (s *AggTradesService) EndTime(endTime int64) *AggTradesService {
	s.endTime = &endTime
	return s
}

// Limit set limit
func (s *AggTradesService) Limit(limit int) *AggTradesService {
	s.limit = &limit
	return s
}

// Do send request
func (s *AggTradesService) Do(ctx context.Context, opts ...RequestOption) (res []*AggTrade, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/aggTrades",
	}
	r.setParam("symbol", s.symbol)
	if s.fromID != nil {
		r.setParam("fromId", *s.fromID)
	}
	if s.startTime != nil {
		r.setParam("startTime", *s.startTime)
	}
	if s.endTime != nil {
		r.setParam("endTime", *s.endTime)
	}
	if s.limit != nil {
		r.setParam("limit", *s.limit)
	}
	data, _, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return []*AggTrade{}, err
	}
	res = make([]*AggTrade, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return []*AggTrade{}, err
	}
	return res, nil
}

// AggTrade define aggregate trade info
type AggTrade struct {
	Price        string `json:"p"`
	Quantity     string `json:"q"`
	AggTradeID   int64  `json:"a"`
	FirstTradeID int64  `json:"f"`
	LastTradeID  int64  `json:"l"`
	Timestamp    int64  `json:"T"`
	IsBuyerMaker bool   `json:"m"`
}

// RecentTradesService list recent trades
type RecentTradesService struct {
	c      *Client
	limit  *int
	symbol string
}

// Symbol set symbol
func (s *RecentTradesService) Symbol(symbol string) *RecentTradesService {
	s.symbol = symbol
	return s
}

// Limit set limit
func (s *RecentTradesService) Limit(limit int) *RecentTradesService {
	s.limit = &limit
	return s
}

// Do send request
func (s *RecentTradesService) Do(ctx context.Context, opts ...RequestOption) (res []*Trade, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/trades",
	}
	r.setParam("symbol", s.symbol)
	if s.limit != nil {
		r.setParam("limit", *s.limit)
	}
	data, _, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return []*Trade{}, err
	}
	res = make([]*Trade, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return []*Trade{}, err
	}
	return res, nil
}

// ListAccountTradeService define account trade list service
type ListAccountTradeService struct {
	c         *Client
	orderId   *int64
	startTime *int64
	endTime   *int64
	fromID    *int64
	limit     *int
	symbol    string
}

// Symbol set symbol
func (s *ListAccountTradeService) Symbol(symbol string) *ListAccountTradeService {
	s.symbol = symbol
	return s
}

// OrderID set orderId
func (s *ListAccountTradeService) OrderID(orderID int64) *ListAccountTradeService {
	s.orderId = &orderID
	return s
}

// StartTime set startTime
func (s *ListAccountTradeService) StartTime(startTime int64) *ListAccountTradeService {
	s.startTime = &startTime
	return s
}

// EndTime set endTime
func (s *ListAccountTradeService) EndTime(endTime int64) *ListAccountTradeService {
	s.endTime = &endTime
	return s
}

// FromID set fromID
func (s *ListAccountTradeService) FromID(fromID int64) *ListAccountTradeService {
	s.fromID = &fromID
	return s
}

// Limit set limit
func (s *ListAccountTradeService) Limit(limit int) *ListAccountTradeService {
	s.limit = &limit
	return s
}

// Do send request
func (s *ListAccountTradeService) Do(ctx context.Context, opts ...RequestOption) (res []*AccountTrade, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/userTrades",
		secType:  secTypeSigned,
	}
	r.setParam("symbol", s.symbol)
	if s.orderId != nil {
		r.setParam("orderId", *s.orderId)
	}
	if s.startTime != nil {
		r.setParam("startTime", *s.startTime)
	}
	if s.endTime != nil {
		r.setParam("endTime", *s.endTime)
	}
	if s.fromID != nil {
		r.setParam("fromID", *s.fromID)
	}
	if s.limit != nil {
		r.setParam("limit", *s.limit)
	}
	data, _, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return []*AccountTrade{}, err
	}
	res = make([]*AccountTrade, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return []*AccountTrade{}, err
	}
	return res, nil
}

// AccountTrade define account trade
type AccountTrade struct {
	RealizedPnl     string           `json:"realizedPnl"`
	Side            SideType         `json:"side"`
	CommissionAsset string           `json:"commissionAsset"`
	Quantity        string           `json:"qty"`
	PositionSide    PositionSideType `json:"positionSide"`
	Price           string           `json:"price"`
	QuoteQuantity   string           `json:"quoteQty"`
	Symbol          string           `json:"symbol"`
	Commission      string           `json:"commission"`
	OrderID         int64            `json:"orderId"`
	ID              int64            `json:"id"`
	Time            int64            `json:"time"`
	Buyer           bool             `json:"buyer"`
	Maker           bool             `json:"maker"`
}
