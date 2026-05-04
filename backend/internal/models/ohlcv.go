package models

import (
	"time"
)

type Timeframe string

const (
	Timeframe1M Timeframe = "1m"
	Timeframe1H Timeframe = "1h"
	Timeframe1D Timeframe = "1d"
)

type OHLCV struct {
	ID        int64     `json:"id"`
	AssetID   int64     `json:"asset_id"`
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	CreatedAt time.Time `json:"created_at"`
}

type OHLCVRequest struct {
	AssetID   int64      `json:"asset_id" query:"asset_id"`
	Timeframe Timeframe  `json:"timeframe" query:"timeframe"`
	From      *time.Time `json:"from,omitempty" query:"from"`
	To        *time.Time `json:"to,omitempty" query:"to"`
	Limit     int        `json:"limit,omitempty" query:"limit"`
}

func (r *OHLCVRequest) SetDefaults() {
	if r.Limit == 0 {
		r.Limit = 100
	}
	if r.Timeframe == "" {
		r.Timeframe = Timeframe1H
	}
}

func (r *OHLCVRequest) TableName() string {
	switch r.Timeframe {
	case Timeframe1M:
		return "ohlcv_1m"
	case Timeframe1H:
		return "ohlcv_1h"
	case Timeframe1D:
		return "ohlcv_1d"
	default:
		return "ohlcv_1h"
	}
}