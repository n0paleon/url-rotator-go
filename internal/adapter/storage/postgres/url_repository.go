package postgres

import (
	"URLRotatorGo/infra/database"
	"URLRotatorGo/infra/logger"
	"URLRotatorGo/internal/core/domain"
	"URLRotatorGo/internal/core/ports"
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"time"
)

type URLRepository struct {
	db *database.Postgres
}

func NewURLRepository(db *database.Postgres) ports.URLRepository {
	return &URLRepository{db}
}

func (r *URLRepository) GetLinks(ctx context.Context, code string) ([]*domain.URL, error) {
	query := r.db.QueryBuilder.Select("id", "shortcode", "total_hit", "original", "created_at", "updated_at").
		From("urls").
		Where(squirrel.Eq{"shortcode": code}).
		OrderBy("RANDOM()")

	sql, args, err := query.ToSql()
	if err != nil {
		logger.L.Errorw("failed to build query", "error", err.Error())
		return nil, domain.ErrInternalServerError
	}

	rows, err := r.db.Pool.Query(ctx, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDataNotFound
		}
		logger.L.Errorw("failed to execute query", "error", err.Error())
		return nil, domain.ErrInternalServerError
	}

	var results []*domain.URL
	for rows.Next() {
		var row domain.URL
		if err = rows.Scan(&row.ID, &row.ShortCode, &row.TotalHit, &row.Original, &row.CreatedAt, &row.UpdatedAt); err != nil {
			logger.L.Errorw("failed to scan row", "error", err.Error())
			return results, domain.ErrInternalServerError
		}
		results = append(results, &row)
	}

	return results, nil
}

func (r *URLRepository) UpdateHit(ctx context.Context, id string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		logger.L.Errorw("failed to start transaction", "error", err.Error())
		return domain.ErrInternalServerError
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	query := r.db.QueryBuilder.Update("urls").
		Set("total_hit", squirrel.Expr("total_hit+1")).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		logger.L.Errorw("failed to build query", "error", err.Error())
		return domain.ErrInternalServerError
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		logger.L.Errorw("failed to execute query", "error", err.Error())
		return domain.ErrInternalServerError
	}

	if err = tx.Commit(ctx); err != nil {
		logger.L.Errorw("failed to commit transaction", "error", err.Error())
		return domain.ErrInternalServerError
	}

	return nil
}

func (r *URLRepository) Save(ctx context.Context, urls []*domain.URL) ([]*domain.URL, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		logger.L.Errorw("failed to create transaction", "error", err.Error())
		return nil, domain.ErrInternalServerError
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	query := r.db.QueryBuilder.Insert("urls").
		Columns("shortcode", "original").
		Suffix("RETURNING *")

	for _, url := range urls {
		query = query.Values(url.ShortCode, url.Original)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		logger.L.Errorw("failed to build query", "error", err.Error())
		return nil, domain.ErrInternalServerError
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		logger.L.Errorw("failed to execute query", "error", err.Error())
		return nil, domain.ErrInternalServerError
	}

	var results []*domain.URL
	for rows.Next() {
		var link domain.URL
		if err = rows.Scan(&link.ID, &link.ShortCode, &link.TotalHit, &link.Original, &link.CreatedAt, &link.UpdatedAt); err != nil {
			logger.L.Errorw("failed to scan row", "error", err.Error())
			return nil, domain.ErrInternalServerError
		}
		results = append(results, &link)
	}

	if err = tx.Commit(ctx); err != nil {
		logger.L.Errorw("failed to commit transaction", "error", err.Error())
		return nil, domain.ErrInternalServerError
	}

	return results, nil
}
