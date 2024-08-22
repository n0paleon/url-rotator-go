package database

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Postgres struct {
	*pgxpool.Pool
	QueryBuilder *squirrel.StatementBuilderType
}

func NewPostgresConn(lc fx.Lifecycle, ctx context.Context, cfg *viper.Viper, log *zap.SugaredLogger) (*Postgres, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.GetString("database.postgres.user"),
		cfg.GetString("database.postgres.pass"),
		cfg.GetString("database.postgres.host"),
		cfg.GetInt("database.postgres.port"),
		cfg.GetString("database.postgres.dbname"),
	)

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err = db.Ping(ctx)
			if err != nil {
				log.Fatalw("postgres connection failed", "error", err.Error())
				return err
			}
			log.Info("postgres connection success")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("postgres connection closed")
			db.Close()
			return nil
		},
	})

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return &Postgres{
		db,
		&psql,
	}, nil
}

func (db *Postgres) Close() {
	db.Pool.Close()
}
