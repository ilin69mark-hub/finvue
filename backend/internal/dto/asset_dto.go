package dto

import "time"

type AssetDTO struct {
	ID               int64      `json:"id"`
	Symbol           string     `json:"symbol"`
	Name             string     `json:"name"`
	AssetType        string     `json:"asset_type"`
	IsActive         bool       `json:"is_active"`
	LastPrice        *float64   `json:"last_price,omitempty"`
	LastPriceUpdated *time.Time `json:"last_price_updated,omitempty"`
}

type CreateAssetDTO struct {
	Symbol    string `json:"symbol" validate:"required,max=20"`
	Name      string `json:"name" validate:"required,max=100"`
	AssetType string `json:"asset_type" validate:"required,oneof=crypto stock forex"`
	IsActive  bool   `json:"is_active"`
}

type UpdateAssetDTO struct {
	Name       string  `json:"name" validate:"max=100"`
	AssetType string  `json:"asset_type" validate:"oneof=crypto stock forex"`
	IsActive   *bool   `json:"is_active,omitempty"`
	LastPrice  float64 `json:"last_price,omitempty"`
}

type AssetListResponse struct {
	Assets   []AssetDTO `json:"assets"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}