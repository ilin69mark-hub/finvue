package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"finvue/internal/models"
	"finvue/internal/pkg/database"
	"finvue/internal/pkg/logger"

	"go.uber.org/zap"
)

type OHLCVRepository struct {
	pool *pgxpool.Pool
}

func NewOHLCVRepository() *OHLCVRepository {
	return &OHLCVRepository{pool: database.Pool}
}

func (r *OHLCVRepository) BatchInsert(ctx context.Context, table string, candles []models.OHLCV) error {
	if len(candles) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (asset_id, timestamp, open, high, low, close, volume, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (asset_id, timestamp) DO UPDATE SET
			open = EXCLUDED.open,
			high = EXCLUDED.high,
			low = EXCLUDED.low,
			close = EXCLUDED.close,
			volume = EXCLUDED.volume
	`, table)

	batch := &pgx.Batch{}
	for _, c := range candles {
		batch.Queue(query, c.AssetID, c.Timestamp, c.Open, c.High, c.Low, c.Close, c.Volume, time.Now())
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(candles); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("ошибка вставки свечи %d: %w", i, err)
		}
	}

	logger.Debug("Вставлено свечей", zap.Int("count", len(candles)), zap.String("table", table))
	return nil
}

func (r *OHLCVRepository) GetByAssetAndTimeframe(ctx context.Context, req models.OHLCVRequest) ([]models.OHLCV, error) {
	req.SetDefaults()
	table := req.TableName()

	query := fmt.Sprintf(`
		SELECT id, asset_id, timestamp, open, high, low, close, volume, created_at
		FROM %s
		WHERE asset_id = $1
	`, table)

	args := []interface{}{req.AssetID}
	argIdx := 2

	if req.From != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIdx)
		args = append(args, *req.From)
		argIdx++
	}

	if req.To != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIdx)
		args = append(args, *req.To)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d", argIdx)
	args = append(args, req.Limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения свечей: %w", err)
	}
	defer rows.Close()

	var candles []models.OHLCV
	for rows.Next() {
		var c models.OHLCV
		err := rows.Scan(&c.ID, &c.AssetID, &c.Timestamp, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume, &c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования свечи: %w", err)
		}
		candles = append(candles, c)
	}

	if candles == nil {
		candles = []models.OHLCV{}
	}

	return candles, nil
}

func (r *OHLCVRepository) GetLatestCandle(ctx context.Context, assetID int64, table string) (*models.OHLCV, error) {
	query := fmt.Sprintf(`
		SELECT id, asset_id, timestamp, open, high, low, close, volume, created_at
		FROM %s
		WHERE asset_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`, table)

	var c models.OHLCV
	err := r.pool.QueryRow(ctx, query, assetID).Scan(
		&c.ID, &c.AssetID, &c.Timestamp, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume, &c.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения последней свечи: %w", err)
	}

	return &c, nil
}

func (r *OHLCVRepository) GetFirstCandle(ctx context.Context, assetID int64, table string) (*models.OHLCV, error) {
	query := fmt.Sprintf(`
		SELECT id, asset_id, timestamp, open, high, low, close, volume, created_at
		FROM %s
		WHERE asset_id = $1
		ORDER BY timestamp ASC
		LIMIT 1
	`, table)

	var c models.OHLCV
	err := r.pool.QueryRow(ctx, query, assetID).Scan(
		&c.ID, &c.AssetID, &c.Timestamp, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume, &c.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения первой свечи: %w", err)
	}

	return &c, nil
}

func (r *OHLCVRepository) DeleteOldCandles(ctx context.Context, table string, before time.Time) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s WHERE timestamp < $1", table)
	result, err := r.pool.Exec(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("ошибка удаления старых свечей: %w", err)
	}

	return result.RowsAffected(), nil
}

func (r *OHLCVRepository) TableExists(ctx context.Context, table string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = $1
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, table).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки существования таблицы: %w", err)
	}

	return exists, nil
}

func (r *OHLCVRepository) GetCandlesInRange(ctx context.Context, table string, assetID int64, from, to time.Time) ([]models.OHLCV, error) {
	query := fmt.Sprintf(`
		SELECT id, asset_id, timestamp, open, high, low, close, volume, created_at
		FROM %s
		WHERE asset_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC
	`, table)

	rows, err := r.pool.Query(ctx, query, assetID, from, to)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения свечей в диапазоне: %w", err)
	}
	defer rows.Close()

	var candles []models.OHLCV
	for rows.Next() {
		var c models.OHLCV
		err := rows.Scan(&c.ID, &c.AssetID, &c.Timestamp, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume, &c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования свечи: %w", err)
		}
		candles = append(candles, c)
	}

	if candles == nil {
		candles = []models.OHLCV{}
	}

	return candles, nil
}

func (r *OHLCVRepository) BuildHigherTimeframeCandle(candles []models.OHLCV) *models.OHLCV {
	if len(candles) == 0 {
		return nil
	}

	open := candles[0].Open
	high := candles[0].High
	low := candles[0].Low
	close := candles[len(candles)-1].Close
	var volume float64

	for _, c := range candles {
		if c.High > high {
			high = c.High
		}
		if c.Low < low {
			low = c.Low
		}
		volume += c.Volume
	}

	return &models.OHLCV{
		Timestamp: candles[0].Timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
	}
}

func (r *OHLCVRepository) AggregateToHigherTimeframe(ctx context.Context, sourceTable, targetTable string, assetID int64) error {
	existing, err := r.GetLatestCandle(ctx, assetID, targetTable)
	if err != nil {
		return err
	}

	var from time.Time
	if existing != nil {
		from = existing.Timestamp
	} else {
		firstCandle, err := r.GetFirstCandle(ctx, assetID, sourceTable)
		if err != nil {
			return err
		}
		if firstCandle == nil {
			return nil
		}
		from = firstCandle.Timestamp
	}

	sourceCandles, err := r.GetCandlesInRange(ctx, sourceTable, assetID, from, time.Now())
	if err != nil {
		return err
	}

	if len(sourceCandles) == 0 {
		return nil
	}

	var targetCandles []models.OHLCV
	intervalStart := sourceCandles[0].Timestamp

	for _, c := range sourceCandles {
		if !c.Timestamp.Before(intervalStart.Add(time.Hour)) {
			if built := r.BuildHigherTimeframeCandle(targetCandles); built != nil {
				built.AssetID = assetID
				targetCandles = append(targetCandles, *built)
			}
			intervalStart = c.Timestamp
			targetCandles = []models.OHLCV{}
		}
		targetCandles = append(targetCandles, c)
	}

	if len(targetCandles) > 0 {
		if built := r.BuildHigherTimeframeCandle(targetCandles); built != nil {
			built.AssetID = assetID
			targetCandles = append(targetCandles, *built)
		}
	}

	if len(targetCandles) > 0 {
		return r.BatchInsert(ctx, targetTable, targetCandles)
	}

	return nil
}

func TableNameFromTimeframe(tf models.Timeframe) string {
	switch tf {
	case models.Timeframe1M:
		return "ohlcv_1m"
	case models.Timeframe1H:
		return "ohlcv_1h"
	case models.Timeframe1D:
		return "ohlcv_1d"
	default:
		return "ohlcv_1h"
	}
}

func ParseTimeframe(tf string) models.Timeframe {
	tf = strings.ToLower(tf)
	switch tf {
	case "1m", "1min", "minute":
		return models.Timeframe1M
	case "1h", "1hour", "hour":
		return models.Timeframe1H
	case "1d", "1day", "day":
		return models.Timeframe1D
	default:
		return models.Timeframe1H
	}
}