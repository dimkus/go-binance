package futures

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// TradeHistoryAsyncDownloadId
// TradeHistoryAsyncDownloadId define position margin history info
type TradeHistoryAsyncDownloadId struct {
	AvgCostTimestampOfLast30D int    `json:"avgCostTimestampOfLast30d"`
	DownloadId                string `json:"downloadId"`
}

// TradeHistoryAsyncDownloadI get position margin history service
type GetTradeHistoryServiceAsyncDownloadId struct {
	c         *Client
	startTime *int64
	endTime   *int64
}

// StartTime set startTime
func (s *GetTradeHistoryServiceAsyncDownloadId) StartTime(startTime int64) *GetTradeHistoryServiceAsyncDownloadId {
	s.startTime = &startTime
	return s
}

// EndTime set endTime
func (s *GetTradeHistoryServiceAsyncDownloadId) EndTime(endTime int64) *GetTradeHistoryServiceAsyncDownloadId {
	s.endTime = &endTime
	return s
}

// Do send request
func (s *GetTradeHistoryServiceAsyncDownloadId) Do(ctx context.Context, opts ...RequestOption) (res *TradeHistoryAsyncDownloadId, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/trades/asyn",
		secType:  secTypeSigned,
	}
	res = new(TradeHistoryAsyncDownloadId)

	if s.startTime == nil {
		return res, errors.New("startTime is not defined")
	}
	r.setParam("startTime", *s.startTime)

	if s.endTime == nil {
		return res, errors.New("endTime is not defined")
	}
	r.setParam("endTime", *s.endTime)

	data, _, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TradeHistoryAsyncDownload
type TradeHistoryAsyncDownload struct {
	DownloadId          string `json:"downloadId"`
	Status              string `json:"status"`
	Url                 string `json:"url"`
	Notified            bool   `json:"notified"`
	ExpirationTimestamp int64  `json:"expirationTimestamp"`
	IsExpired           bool   `json:"isExpired"`
}

// TradeHistoryAsyncDownload get position margin history service
type GetTradeHistoryAsyncDownload struct {
	c          *Client
	downloadId string
}

// DownloadId set downloadId
func (s *GetTradeHistoryAsyncDownload) DownloadId(downloadId string) *GetTradeHistoryAsyncDownload {
	s.downloadId = downloadId
	return s
}

// Do send request
func (s *GetTradeHistoryAsyncDownload) Do(ctx context.Context, opts ...RequestOption) (res *TradeHistoryAsyncDownload, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/trade/asyn/id",
		secType:  secTypeSigned,
	}
	res = new(TradeHistoryAsyncDownload)

	if s.downloadId == "" {
		return res, errors.New("downloadId is not defined")
	}
	r.setParam("downloadId", s.downloadId)

	data, _, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
