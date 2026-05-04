package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"finvue/internal/models"
	"finvue/internal/pkg/database"
	"finvue/internal/pkg/logger"

	"go.uber.org/zap"
)

type AssetRepository struct {
	pool *pgxpool.Pool
}

func NewAssetRepository() *AssetRepository {
	return &AssetRepository{pool: database.Pool}
}

func (r *AssetRepository) Create(ctx context.Context, asset *models.Asset) error {
	query := `
		INSERT INTO assets (symbol, name, asset_type, is_active, last_price, last_price_updated, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	var id int64
	var createdAt, updatedAt time.Time

	err := r.pool.QueryRow(ctx, query,
		asset.Symbol,
		asset.Name,
		asset.AssetType,
		asset.IsActive,
		asset.LastPrice,
		asset.LastPriceUpdated,
		time.Now(),
		time.Now(),
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return fmt.Errorf("ошибка создания актива: %w", err)
	}

	asset.ID = id
	asset.CreatedAt = createdAt
	asset.UpdatedAt = updatedAt

	logger.Debug("Актив создан", zap.Int64("id", id), zap.String("symbol", asset.Symbol))

	return nil
}

func (r *AssetRepository) GetByID(ctx context.Context, id int64) (*models.Asset, error) {
	query := `
		SELECT id, symbol, name, asset_type, is_active, last_price, last_price_updated, created_at, updated_at
		FROM assets
		WHERE id = $1
	`

	var asset models.Asset
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&asset.ID,
		&asset.Symbol,
		&asset.Name,
		&asset.AssetType,
		&asset.IsActive,
		&asset.LastPrice,
		&asset.LastPriceUpdated,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения актива: %w", err)
	}

	return &asset, nil
}

func (r *AssetRepository) GetBySymbol(ctx context.Context, symbol string) (*models.Asset, error) {
	query := `
		SELECT id, symbol, name, asset_type, is_active, last_price, last_price_updated, created_at, updated_at
		FROM assets
		WHERE symbol = $1
	`

	var asset models.Asset
	err := r.pool.QueryRow(ctx, query, symbol).Scan(
		&asset.ID,
		&asset.Symbol,
		&asset.Name,
		&asset.AssetType,
		&asset.IsActive,
		&asset.LastPrice,
		&asset.LastPriceUpdated,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения актива по символу: %w", err)
	}

	return &asset, nil
}

func (r *AssetRepository) GetAll(ctx context.Context, includeInactive bool) ([]models.Asset, error) {
	var query string
	var args []interface{}

	if includeInactive {
		query = `
			SELECT id, symbol, name, asset_type, is_active, last_price, last_price_updated, created_at, updated_at
			FROM assets
			ORDER BY symbol ASC
		`
	} else {
		query = `
			SELECT id, symbol, name, asset_type, is_active, last_price, last_price_updated, created_at, updated_at
			FROM assets
			WHERE is_active = true
			ORDER BY symbol ASC
		`
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка активов: %w", err)
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var asset models.Asset
		err := rows.Scan(
			&asset.ID,
			&asset.Symbol,
			&asset.Name,
			&asset.AssetType,
			&asset.IsActive,
			&asset.LastPrice,
			&asset.LastPriceUpdated,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования актива: %w", err)
		}
		assets = append(assets, asset)
	}

	return assets, nil
}

func (r *AssetRepository) Update(ctx context.Context, asset *models.Asset) error {
	query := `
		UPDATE assets
		SET name = $1, asset_type = $2, is_active = $3, last_price = $4, last_price_updated = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.pool.Exec(ctx, query,
		asset.Name,
		asset.AssetType,
		asset.IsActive,
		asset.LastPrice,
		asset.LastPriceUpdated,
		time.Now(),
		asset.ID,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления актива: %w", err)
	}

	asset.UpdatedAt = time.Now()
	return nil
}

func (r *AssetRepository) UpsertFromSymbol(ctx context.Context, symbol, name string, assetType models.AssetType) (*models.Asset, error) {
	existing, err := r.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return existing, nil
	}

	asset := &models.Asset{
		Symbol:    symbol,
		Name:      name,
		AssetType: assetType,
		IsActive:  true,
	}

	if err := r.Create(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}