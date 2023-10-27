package futures

import (
	"context"
	"encoding/json"
	"net/http"
)

// TopLongShortAccountRatioService list open history data of a symbol.
type TopLongShortAccountRatioService struct {
	c         *Client
	symbol    string
	period    string
	limit     *int
	startTime *int64
	endTime   *int64
}

// Symbol set symbol
func (s *TopLongShortAccountRatioService) Symbol(symbol string) *TopLongShortAccountRatioService {
	s.symbol = symbol
	return s
}

// Period set period interval
func (s *TopLongShortAccountRatioService) Period(period string) *TopLongShortAccountRatioService {
	s.period = period
	return s
}

// Limit set limit
func (s *TopLongShortAccountRatioService) Limit(limit int) *TopLongShortAccountRatioService {
	s.limit = &limit
	return s
}

// StartTime set startTime
func (s *TopLongShortAccountRatioService) StartTime(startTime int64) *TopLongShortAccountRatioService {
	s.startTime = &startTime
	return s
}

// EndTime set endTime
func (s *TopLongShortAccountRatioService) EndTime(endTime int64) *TopLongShortAccountRatioService {
	s.endTime = &endTime
	return s
}

// Do send request
func (s *TopLongShortAccountRatioService) Do(ctx context.Context, opts ...RequestOption) (res []*TopLongShortAccountRatio, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/futures/data/topLongShortAccountRatio",
	}

	r.setParam("symbol", s.symbol)
	r.setParam("period", s.period)

	if s.limit != nil {
		r.setParam("limit", *s.limit)
	}
	if s.startTime != nil {
		r.setParam("startTime", *s.startTime)
	}
	if s.endTime != nil {
		r.setParam("endTime", *s.endTime)
	}

	data, _, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return []*TopLongShortAccountRatio{}, err
	}

	res = make([]*TopLongShortAccountRatio, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return []*TopLongShortAccountRatio{}, err
	}

	return res, nil
}

type TopLongShortAccountRatio struct {
	Symbol         string `json:"symbol"`
	LongShortRatio string `json:"longShortRatio"`
	LongAccount    string `json:"longAccount"`
	ShortAccount   string `json:"shortAccount"`
	Timestamp      int64  `json:"timestamp"`
}
