package models

import (
	"time"
)

type AlertType string

const (
	AlertTypeSMACrossover AlertType = "sma_crossover"
	AlertTypePriceAbove   AlertType = "price_above"
	AlertTypePriceBelow   AlertType = "price_below"
	AlertTypeVolumeSpike  AlertType = "volume_spike"
	AlertTypePriceChange  AlertType = "price_change"
)

type Alert struct {
	ID        int64     `json:"id"`
	AssetID   int64     `json:"asset_id"`
	Asset     *Asset    `json:"asset,omitempty"`
	AlertType AlertType `json:"alert_type"`
	Message   string    `json:"message"`
	Value     float64   `json:"value,omitempty"`
	Threshold float64   `json:"threshold,omitempty"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *Alert) SetRead() {
	a.IsRead = true
}

func (a *Alert) SetUnread() {
	a.IsRead = false
}

type AlertCreate struct {
	AssetID   int64     `json:"asset_id"`
	AlertType AlertType `json:"alert_type"`
	Message   string    `json:"message"`
	Value     float64   `json:"value,omitempty"`
	Threshold float64   `json:"threshold,omitempty"`
}
