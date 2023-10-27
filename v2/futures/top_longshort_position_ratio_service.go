package futures

import (
	"context"
	"encoding/json"
	"net/http"
)

// TopLongShortPositionRatioService list open history data of a symbol.
type TopLongShortPositionRatioService struct {
	c         *Client
	symbol    string
	period    string
	limit     *int
	startTime *int64
	endTime   *int64
}

// Symbol set symbol
func (s *TopLongShortPositionRatioService) Symbol(symbol string) *TopLongShortPositionRatioService {
	s.symbol = symbol
	return s
}

// Period set period interval
func (s *TopLongShortPositionRatioService) Period(period string) *TopLongShortPositionRatioService {
	s.period = period
	return s
}

// Limit set limit
func (s *TopLongShortPositionRatioService) Limit(limit int) *TopLongShortPositionRatioService {
	s.limit = &limit
	return s
}

// StartTime set startTime
func (s *TopLongShortPositionRatioService) StartTime(startTime int64) *TopLongShortPositionRatioService {
	s.startTime = &startTime
	return s
}

// EndTime set endTime
func (s *TopLongShortPositionRatioService) EndTime(endTime int64) *TopLongShortPositionRatioService {
	s.endTime = &endTime
	return s
}

// Do send request
func (s *TopLongShortPositionRatioService) Do(ctx context.Context, opts ...RequestOption) (res []*TopLongShortPositionRatio, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/futures/data/topLongShortPositionRatio",
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
		return []*TopLongShortPositionRatio{}, err
	}

	res = make([]*TopLongShortPositionRatio, 0)
	err = json.Unmarshal(data, &res)
	if err != nil {
		return []*TopLongShortPositionRatio{}, err
	}

	return res, nil
}

type TopLongShortPositionRatio struct {
	Symbol         string `json:"symbol"`
	LongShortRatio string `json:"longShortRatio"`
	LongAccount    string `json:"longAccount"`
	ShortAccount   string `json:"shortAccount"`
	Timestamp      int64  `json:"timestamp"`
}
