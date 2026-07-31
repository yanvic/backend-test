package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/autoparts/backend-test/internal/domain"
	"github.com/autoparts/backend-test/internal/repository"
)

type Repo struct {
	mu sync.RWMutex
	db *sql.DB
}

var _ repository.PartRepository = (*Repo)(nil)

func New(dbPath string) (*Repo, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &Repo{db: db}, nil
}

func migrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS parts (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		category TEXT NOT NULL DEFAULT '',
		current_stock INTEGER NOT NULL DEFAULT 0,
		minimum_stock INTEGER NOT NULL DEFAULT 0,
		average_daily_sales REAL NOT NULL DEFAULT 0,
		lead_time_days INTEGER NOT NULL DEFAULT 0,
		unit_cost REAL NOT NULL DEFAULT 0,
		criticality_level INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	`
	_, err := db.Exec(query)
	return err
}

func (r *Repo) Create(_ context.Context, part *domain.Part) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(
		`INSERT INTO parts (id, name, category, current_stock, minimum_stock, average_daily_sales, lead_time_days, unit_cost, criticality_level, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		part.ID, part.Name, part.Category, int(part.CurrentStock), int(part.MinimumStock),
		float64(part.AverageDailySales), int(part.LeadTimeDays), part.UnitCost, int(part.CriticalityLevel),
		part.CreatedAt.Format(time.RFC3339), part.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint") {
		return fmt.Errorf(domain.MsgPartExists, part.ID)
	}
	return err
}

func (r *Repo) GetByID(_ context.Context, id string) (*domain.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	row := r.db.QueryRow(
		`SELECT id, name, category, current_stock, minimum_stock, average_daily_sales, lead_time_days, unit_cost, criticality_level, created_at, updated_at
		 FROM parts WHERE id = ?`, id,
	)
	part, err := scanPart(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(domain.MsgPartNotFound, id)
		}
		return nil, err
	}
	return part, nil
}

func (r *Repo) GetAll(_ context.Context) ([]*domain.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rows, err := r.db.Query(
		`SELECT id, name, category, current_stock, minimum_stock, average_daily_sales, lead_time_days, unit_cost, criticality_level, created_at, updated_at
		 FROM parts`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parts := make([]*domain.Part, 0)
	for rows.Next() {
		part, err := scanPart(rows)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, rows.Err()
}

func (r *Repo) Update(_ context.Context, part *domain.Part) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result, err := r.db.Exec(
		`UPDATE parts SET name=?, category=?, current_stock=?, minimum_stock=?, average_daily_sales=?, lead_time_days=?, unit_cost=?, criticality_level=?, updated_at=?
		 WHERE id=?`,
		part.Name, part.Category, int(part.CurrentStock), int(part.MinimumStock),
		float64(part.AverageDailySales), int(part.LeadTimeDays), part.UnitCost, int(part.CriticalityLevel),
		part.UpdatedAt.Format(time.RFC3339), part.ID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf(domain.MsgPartNotFound, part.ID)
	}
	return nil
}

func (r *Repo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result, err := r.db.Exec(`DELETE FROM parts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf(domain.MsgPartNotFound, id)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanPart(s rowScanner) (*domain.Part, error) {
	var part domain.Part
	var createdAt, updatedAt string

	err := s.Scan(
		&part.ID, &part.Name, &part.Category,
		(*int)(&part.CurrentStock), (*int)(&part.MinimumStock),
		(*float64)(&part.AverageDailySales), (*int)(&part.LeadTimeDays),
		&part.UnitCost,
		(*int)(&part.CriticalityLevel),
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	part.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	part.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &part, nil
}
