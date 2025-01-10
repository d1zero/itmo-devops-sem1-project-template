package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DbName   string
}

type Transactor interface {
	InTx(ctx context.Context, f func(ctx context.Context) error) error
}

type DB struct {
	pool *pgxpool.Pool
	sq   squirrel.StatementBuilderType
}

func New(ctx context.Context, cfg Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.DbName,
		30,
		60*time.Second,
		60*time.Second,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}

	db := &DB{
		pool: pool,
		sq:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return db, nil
}

type contextKey string

const txContextKey contextKey = "postgres_tx"

func (db *DB) InTx(ctx context.Context, f func(ctx context.Context) error) error {
	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	ctx = context.WithValue(ctx, txContextKey, tx)

	if err = f(ctx); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (db *DB) Builder() squirrel.StatementBuilderType {
	return db.sq
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db *DB) exec(ctx context.Context, qb squirrel.Sqlizer) (pgconn.CommandTag, error) {
	sql, args, err := qb.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("failed to cast query to sql: %w", err)
	}

	ct, err := db.queryRunner(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("postgres.DB.Exec: %w", err)
	}

	return ct, nil
}

func (db *DB) Insert(ctx context.Context, qb squirrel.InsertBuilder) (pgconn.CommandTag, error) {
	return db.exec(ctx, qb)
}

func (db *DB) selectData(ctx context.Context, dest any, qb squirrel.Sqlizer) error {
	sql, args, err := qb.ToSql()
	if err != nil {
		return fmt.Errorf("failed to cast query to sql: %w", err)
	}

	if err = pgxscan.Select(ctx, db.queryRunner(ctx), dest, sql, args...); err != nil {
		return fmt.Errorf("postgres.DB.Select: %w", err)
	}

	return nil
}

func (db *DB) Select(ctx context.Context, dest any, qb squirrel.SelectBuilder) error {
	return db.selectData(ctx, dest, qb)
}

type queryRunner interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (db *DB) queryRunner(ctx context.Context) queryRunner {
	tx, ok := ctx.Value(txContextKey).(pgx.Tx)
	if !ok {
		return db.pool
	}

	return tx
}
