package storage

import (
	"context"
	"fmt"
	"time"

	"L3_6/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}

func (s *Storage) CreateSale(sale *models.Sale) error {
	const op = "storage.CreateSale"

	query := `INSERT INTO sales (type, amount, date, category) VALUES ($1, $2, $3, $4) RETURNING id`
	err := s.db.QueryRow(context.Background(), query, sale.Type, sale.Amount, sale.Date, sale.Category).Scan(&sale.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetSales() ([]models.Sale, error) {
	const op = "storage.GetSales"

	query := `SELECT id, type, amount, date, category FROM sales ORDER BY date DESC`
	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var sales []models.Sale
	for rows.Next() {
		var sale models.Sale
		err := rows.Scan(&sale.ID, &sale.Type, &sale.Amount, &sale.Date, &sale.Category)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		sales = append(sales, sale)
	}

	return sales, nil
}

func (s *Storage) UpdateSale(sale *models.Sale) error {
	const op = "storage.UpdateSale"

	query := `UPDATE sales SET type=$1, amount=$2, date=$3, category=$4 WHERE id=$5`
	_, err := s.db.Exec(context.Background(), query, sale.Type, sale.Amount, sale.Date, sale.Category, sale.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteSale(id int) error {
	const op = "storage.DeleteSale"

	query := `DELETE FROM sales WHERE id=$1`
	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetAnalytics(from, to time.Time) (*models.AnalyticsResponse, error) {
	const op = "storage.GetAnalytics"

	query := `
		SELECT 
			COALESCE(SUM(amount), 0) as sum,
			COALESCE(AVG(amount), 0) as average,
			COUNT(*) as count,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount) as median,
			PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY amount) as percentile90
		FROM sales 
		WHERE date BETWEEN $1 AND $2
	`

	var analytics models.AnalyticsResponse
	err := s.db.QueryRow(context.Background(), query, from, to).Scan(
		&analytics.Sum,
		&analytics.Average,
		&analytics.Count,
		&analytics.Median,
		&analytics.Percentile90,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &analytics, nil
}
