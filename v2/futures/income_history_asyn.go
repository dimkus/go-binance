package futures

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// IncomeHistoryAyncDownloadId
type IncomeHistoryAsyncDownloadId struct {
	AvgCostTimestampOfLast30D int    `json:"avgCostTimestampOfLast30d"`
	DownloadId                string `json:"downloadId"`
}

// IncomeHistoryAyncDownloadId get position margin history service
type GetIncomeHistoryServiceAsyncDownloadId struct {
	c         *Client
	startTime *int64
	endTime   *int64
}

// StartTime set startTime
func (s *GetIncomeHistoryServiceAsyncDownloadId) StartTime(startTime int64) *GetIncomeHistoryServiceAsyncDownloadId {
	s.startTime = &startTime
	return s
}

// EndTime set endTime
func (s *GetIncomeHistoryServiceAsyncDownloadId) EndTime(endTime int64) *GetIncomeHistoryServiceAsyncDownloadId {
	s.endTime = &endTime
	return s
}

// Do send request
func (s *GetIncomeHistoryServiceAsyncDownloadId) Do(ctx context.Context, opts ...RequestOption) (res *IncomeHistoryAsyncDownloadId, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/income/asyn",
		secType:  secTypeSigned,
	}
	res = new(IncomeHistoryAsyncDownloadId)

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

// IncomeHistoryAsyncDownload
type IncomeHistoryAsyncDownload struct {
	DownloadId          string `json:"downloadId"`
	Status              string `json:"status"`
	Url                 string `json:"url"`
	Notified            bool   `json:"notified"`
	ExpirationTimestamp int64  `json:"expirationTimestamp"`
	IsExpired           bool   `json:"isExpired"`
}

// IncomeHistoryAsyncDownload get position margin history service
type GetIncomeHistoryAsyncDownload struct {
	c          *Client
	downloadId string
}

// DownloadId set downloadId
func (s *GetIncomeHistoryAsyncDownload) DownloadId(downloadId string) *GetIncomeHistoryAsyncDownload {
	s.downloadId = downloadId
	return s
}

// Do send request
func (s *GetIncomeHistoryAsyncDownload) Do(ctx context.Context, opts ...RequestOption) (res *IncomeHistoryAsyncDownload, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/fapi/v1/income/asyn/id",
		secType:  secTypeSigned,
	}
	res = new(IncomeHistoryAsyncDownload)

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
