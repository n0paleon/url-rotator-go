package database

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Redis struct {
	*redis.Client
}

func NewRedisConn(lc fx.Lifecycle, ctx context.Context, cfg *viper.Viper, log *zap.SugaredLogger) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.GetString("database.redis.addr"),
		Password: cfg.GetString("database.redis.passwd"),
		DB:       cfg.GetInt("database.redis.db"),
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := rdb.Ping(ctx).Err()
			if err != nil {
				log.Fatalw("Redis connection error", "error", err.Error())
				return err
			}
			log.Info("Redis connection success")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Redis connection closed")
			return rdb.Close()
		},
	})

	return &Redis{rdb}
}

func (r *Redis) Close() error {
	return r.Client.Close()
}
