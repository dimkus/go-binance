package futures

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// OrderHistoryAsyncDownloadId
type OrderHistoryAsyncDownloadId struct {
	AvgCostTimestampOfLast30D int    `json:"avgCostTimestampOfLast30d"`
	DownloadId                string `json:"downloadId"`
}

// GetOrderHistoryServiceAsyncDownloadId
type GetOrderHistoryServiceAsyncDownloadId struct {
	c         *Client
	startTime *int64
	endTime   *int64
}

// StartTime set startTime
func (s *GetOrderHistoryServiceAsyncDownloadId) StartTime(startTime int64) *GetOrderHistoryServiceAsyncDownloadId {
	s.startTime = &startTime
	return s
}

// EndTime set endTime
func (s *GetOrderHistoryServiceAsyncDownloadId) EndTime(endTime int64) *GetOrderHistoryServiceAsyncDownloadId {
	s.endTime = &endTime
	return s
}

// Do send request
func (s *GetOrderHistoryServiceAsyncDownloadId) Do(ctx context.Context, opts ...RequestOption) (res *OrderHistoryAsyncDownloadId, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/order/asyn",
		secType:  secTypeSigned,
	}
	res = new(OrderHistoryAsyncDownloadId)

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

// OrderHistoryAsyncDownload
type OrderHistoryAsyncDownload struct {
	DownloadId          string `json:"downloadId"`
	Status              string `json:"status"`
	Url                 string `json:"url"`
	Notified            bool   `json:"notified"`
	ExpirationTimestamp int64  `json:"expirationTimestamp"`
	IsExpired           bool   `json:"isExpired"`
}

// OrderHistoryAsyncDownload get position margin history service
type GetOrderHistoryAsyncDownload struct {
	c          *Client
	downloadId string
}

// DownloadId set downloadId
func (s *GetOrderHistoryAsyncDownload) DownloadId(downloadId string) *GetOrderHistoryAsyncDownload {
	s.downloadId = downloadId
	return s
}

// Do send request
func (s *GetOrderHistoryAsyncDownload) Do(ctx context.Context, opts ...RequestOption) (res *OrderHistoryAsyncDownload, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/order/asyn/id",
		secType:  secTypeSigned,
	}
	res = new(OrderHistoryAsyncDownload)

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
