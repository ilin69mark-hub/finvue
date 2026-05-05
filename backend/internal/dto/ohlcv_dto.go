package dto

import "time"

type OHLCVDTO struct {
	ID        int64   `json:"id"`
	AssetID   int64   `json:"asset_id"`
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

type OHLCVListResponse struct {
	Asset     AssetDTO   `json:"asset"`
	Timeframe string     `json:"timeframe"`
	Candles   []OHLCVDTO `json:"candles"`
	Total     int        `json:"total"`
}

type OHLCVQueryDTO struct {
	AssetID   int64      `json:"asset_id" validate:"required"`
	Timeframe string     `json:"timeframe" validate:"oneof=1m 1h 1d"`
	From      *time.Time `json:"from,omitempty"`
	To        *time.Time `json:"to,omitempty"`
	Limit     int        `json:"limit" validate:"min=1,max=1000"`
}

func (q *OHLCVQueryDTO) SetDefaults() {
	if q.Timeframe == "" {
		q.Timeframe = "1h"
	}
	if q.Limit == 0 {
		q.Limit = 100
	}
}
