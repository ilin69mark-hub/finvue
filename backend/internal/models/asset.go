package models

import (
	"time"
)

type AssetType string

const (
	AssetTypeCrypto AssetType = "crypto"
	AssetTypeStock  AssetType = "stock"
	AssetTypeForex  AssetType = "forex"
)

type Asset struct {
	ID               int64      `json:"id"`
	Symbol           string     `json:"symbol"`
	Name             string     `json:"name"`
	AssetType        AssetType  `json:"asset_type"`
	IsActive         bool       `json:"is_active"`
	LastPrice        float64    `json:"last_price,omitempty"`
	LastPriceUpdated *time.Time `json:"last_price_updated,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (a *Asset) SetLastPrice(price float64) {
	a.LastPrice = price
	now := time.Now()
	a.LastPriceUpdated = &now
}

func (a *Asset) GetLastPrice() float64 {
	return a.LastPrice
}

func (a *Asset) IsActive_Getter() bool {
	return a.IsActive
}

func (at AssetType) String() string {
	return string(at)
}
