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

type ShortCodeRepository struct {
	db *database.Postgres
}

func NewShortCodeRepository(db *database.Postgres) ports.ShortCodeRepository {
	return &ShortCodeRepository{db}
}

func (r *ShortCodeRepository) GetShortCode(ctx context.Context, code string) (*domain.ShortCode, error) {
	query := r.db.QueryBuilder.Select("id", "code", "total_hit", "strategy", "created_at", "updated_at").
		From("shortcodes").
		Where(squirrel.Eq{"code": code}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		logger.L.Errorw("failed to build query", "error", err.Error())
		return nil, domain.ErrInternalServerError
	}

	var data domain.ShortCode
	err = r.db.Pool.QueryRow(ctx, sql, args...).
		Scan(
			&data.ID,
			&data.Code,
			&data.TotalHit,
			&data.Strategy,
			&data.CreatedAt,
			&data.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDataNotFound
		}

		logger.L.Errorw("failed to execute query", "error", err.Error())
		return nil, domain.ErrInternalServerError
	}

	return &data, nil
}

func (r *ShortCodeRepository) UpdateHit(ctx context.Context, code string) error {
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

	query := r.db.QueryBuilder.Update("shortcodes").
		Set("total_hit", squirrel.Expr("total_hit+1")).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"code": code})

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

func (r *ShortCodeRepository) Save(ctx context.Context, url *domain.ShortCode) (*domain.ShortCode, error) {
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

	query := r.db.QueryBuilder.Insert("shortcodes").
		Columns("code", "strategy").
		Values(url.Code, url.Strategy).
		Suffix("RETURNING id, code, strategy, created_at")

	sql, args, err := query.ToSql()
	if err != nil {
		logger.L.Errorw("failed to build query", "error", err.Error())
		return nil, domain.ErrInternalServerError
	}

	err = tx.QueryRow(ctx, sql, args...).Scan(&url.ID, &url.Code, &url.Strategy, &url.CreatedAt)
	if err != nil {
		logger.L.Errorw("failed to insert shortcode", "error", err.Error())
		return nil, domain.ErrInternalServerError
	}

	if err = tx.Commit(ctx); err != nil {
		logger.L.Errorw("failed to commit transaction", "error", err.Error())
		return nil, domain.ErrInternalServerError
	}

	return url, nil
}
