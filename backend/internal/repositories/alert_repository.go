package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"finvue/internal/models"
	"finvue/internal/pkg/database"
)

type AlertRepository struct {
	pool *pgxpool.Pool
}

func NewAlertRepository() *AlertRepository {
	return &AlertRepository{pool: database.Pool}
}

func (r *AlertRepository) Create(ctx context.Context, alert *models.Alert) error {
	query := `
		INSERT INTO alerts (asset_id, alert_type, message, value, threshold, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	var id int64
	var createdAt time.Time

	err := r.pool.QueryRow(ctx, query,
		alert.AssetID,
		alert.AlertType,
		alert.Message,
		alert.Value,
		alert.Threshold,
		alert.IsRead,
		time.Now(),
	).Scan(&id, &createdAt)

	if err != nil {
		return fmt.Errorf("ошибка создания алерта: %w", err)
	}

	alert.ID = id
	alert.CreatedAt = createdAt

	return nil
}

func (r *AlertRepository) GetAll(ctx context.Context, unreadOnly bool) ([]models.Alert, error) {
	var query string
	if unreadOnly {
		query = `
			SELECT id, asset_id, alert_type, message, value, threshold, is_read, created_at
			FROM alerts
			WHERE is_read = false
			ORDER BY created_at DESC
		`
	} else {
		query = `
			SELECT id, asset_id, alert_type, message, value, threshold, is_read, created_at
			FROM alerts
			ORDER BY created_at DESC
		`
	}

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения алертов: %w", err)
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		err := rows.Scan(
			&alert.ID,
			&alert.AssetID,
			&alert.AlertType,
			&alert.Message,
			&alert.Value,
			&alert.Threshold,
			&alert.IsRead,
			&alert.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования алерта: %w", err)
		}
		alerts = append(alerts, alert)
	}

	if alerts == nil {
		alerts = []models.Alert{}
	}

	return alerts, nil
}

func (r *AlertRepository) MarkRead(ctx context.Context, id int64) error {
	query := `UPDATE alerts SET is_read = true WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *AlertRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM alerts WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *AlertRepository) GetLastByAssetAndType(ctx context.Context, assetID int64, alertType models.AlertType) (*models.Alert, error) {
	query := `
		SELECT id, asset_id, alert_type, message, value, threshold, is_read, created_at
		FROM alerts
		WHERE asset_id = $1 AND alert_type = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var alert models.Alert
	err := r.pool.QueryRow(ctx, query, assetID, alertType).Scan(
		&alert.ID,
		&alert.AssetID,
		&alert.AlertType,
		&alert.Message,
		&alert.Value,
		&alert.Threshold,
		&alert.IsRead,
		&alert.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &alert, nil
}