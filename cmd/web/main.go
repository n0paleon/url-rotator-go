package main

import (
	"URLRotatorGo/infra/config"
	"URLRotatorGo/infra/database"
	"URLRotatorGo/infra/httpserver"
	"URLRotatorGo/infra/logger"
	"URLRotatorGo/infra/workerpool"
	"URLRotatorGo/internal/adapter/http"
	"URLRotatorGo/internal/adapter/http/handler"
	"URLRotatorGo/internal/adapter/storage/cache"
	"URLRotatorGo/internal/adapter/storage/postgres"
	"URLRotatorGo/internal/core/ports"
	"URLRotatorGo/internal/core/services"
	"context"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	cfg := config.InitConfig("config.json", "../..")
	mylog := logger.NewLogger(cfg)
	workerpool.IntializePool(cfg, mylog)

	ctx := context.Background()

	fx.New(
		fx.Provide(func() *viper.Viper { return cfg }),
		fx.Provide(func() *zap.SugaredLogger { return mylog }),
		fx.Provide(func() context.Context { return ctx }),
		fx.Provide(httpserver.InitServer),
		fx.Provide(database.NewPostgresConn),
		fx.Provide(database.NewRedisConn),
		fx.Provide(
			fx.Annotate(
				postgres.NewShortCodeRepository,
				fx.As(new(ports.ShortCodeRepository)),
			),
			fx.Annotate(
				postgres.NewURLRepository,
				fx.As(new(ports.URLRepository)),
			),
		),
		fx.Provide(
			fx.Annotate(
				cache.NewRedisCache,
				fx.As(new(ports.CacheRepository)),
			),
		),
		fx.Provide(
			fx.Annotate(
				services.NewShortenerService,
				fx.As(new(ports.ShortenerService)),
			),
		),
		fx.Provide(
			handler.NewURLHandler,
			http.NewRouter,
		),
		fx.Invoke(func(r *http.Router) {
			r.SetupRoutes()
		}),
	).Run()
}
