package services

import (
	"context"

	"finvue/internal/dto"
	"finvue/internal/models"
	"finvue/internal/repositories"
)

type AssetService struct {
	repo *repositories.AssetRepository
}

func NewAssetService(repo *repositories.AssetRepository) *AssetService {
	return &AssetService{repo: repo}
}

func (s *AssetService) GetAll(ctx context.Context, includeInactive bool) ([]dto.AssetDTO, error) {
	assets, err := s.repo.GetAll(ctx, includeInactive)
	if err != nil {
		return nil, err
	}

	dtos := make([]dto.AssetDTO, 0, len(assets))
	for _, a := range assets {
		dtos = append(dtos, s.modelToDTO(a))
	}

	return dtos, nil
}

func (s *AssetService) GetByID(ctx context.Context, id int64) (*dto.AssetDTO, error) {
	asset, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, nil
	}

	dto := s.modelToDTO(*asset)
	return &dto, nil
}

func (s *AssetService) GetBySymbol(ctx context.Context, symbol string) (*dto.AssetDTO, error) {
	asset, err := s.repo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, nil
	}

	dto := s.modelToDTO(*asset)
	return &dto, nil
}

func (s *AssetService) Create(ctx context.Context, input dto.CreateAssetDTO) (*dto.AssetDTO, error) {
	assetType := models.AssetType(input.AssetType)
	if assetType == "" {
		assetType = models.AssetTypeCrypto
	}

	asset := &models.Asset{
		Symbol:    input.Symbol,
		Name:      input.Name,
		AssetType: assetType,
		IsActive:  input.IsActive,
	}

	if err := s.repo.Create(ctx, asset); err != nil {
		return nil, err
	}

	dto := s.modelToDTO(*asset)
	return &dto, nil
}

func (s *AssetService) Update(ctx context.Context, id int64, input dto.UpdateAssetDTO) (*dto.AssetDTO, error) {
	asset, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, nil
	}

	if input.Name != "" {
		asset.Name = input.Name
	}
	if input.AssetType != "" {
		asset.AssetType = models.AssetType(input.AssetType)
	}
	if input.IsActive != nil {
		asset.IsActive = *input.IsActive
	}
	if input.LastPrice > 0 {
		asset.SetLastPrice(input.LastPrice)
	}

	if err := s.repo.Update(ctx, asset); err != nil {
		return nil, err
	}

	dto := s.modelToDTO(*asset)
	return &dto, nil
}

func (s *AssetService) UpsertFromSymbol(ctx context.Context, symbol, name string, assetType string) (*dto.AssetDTO, error) {
	at := models.AssetType(assetType)
	if at == "" {
		at = models.AssetTypeCrypto
	}

	asset, err := s.repo.UpsertFromSymbol(ctx, symbol, name, at)
	if err != nil {
		return nil, err
	}

	dto := s.modelToDTO(*asset)
	return &dto, nil
}

func (s *AssetService) modelToDTO(asset models.Asset) dto.AssetDTO {
	dto := dto.AssetDTO{
		ID:        asset.ID,
		Symbol:    asset.Symbol,
		Name:      asset.Name,
		AssetType: string(asset.AssetType),
		IsActive:  asset.IsActive,
	}

	if asset.LastPrice > 0 {
		dto.LastPrice = &asset.LastPrice
	}

	if asset.LastPriceUpdated != nil {
		dto.LastPriceUpdated = asset.LastPriceUpdated
	}

	return dto
}