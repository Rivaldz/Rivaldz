// Package postgres implements postgres connection.
package postgres

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres -.
type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
}

// New -.
func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = safeIntToInt32(pg.maxPoolSize)

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - connAttempts == 0: %w", err)
	}

	return pg, nil
}

// Close -.
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

// Transaction executes fn within a database transaction and commits it.
// The transaction is rolled back if fn returns an error.
func (p *Postgres) Transaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres - Transaction - p.Pool.Begin: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx) //nolint:errcheck // rollback after commit is a benign no-op
	}()

	if err = fn(tx); err != nil {
		return fmt.Errorf("postgres - Transaction - fn: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres - Transaction - tx.Commit: %w", err)
	}

	return nil
}

func safeIntToInt32(v int) int32 {
	clamped := v & math.MaxInt32

	return int32(clamped)
}
