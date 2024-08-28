package httpserver

import (
	"URLRotatorGo/infra/logger"
	"URLRotatorGo/infra/workerpool"
	"URLRotatorGo/pkg"
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"
	"time"
)

func InitServer(lc fx.Lifecycle, cfg *viper.Viper) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:            cfg.GetString("app.name"),
		CaseSensitive:      true,
		EnablePrintRoutes:  true,
		JSONEncoder:        sonic.Marshal,
		JSONDecoder:        sonic.Unmarshal,
		StrictRouting:      true,
		WriteTimeout:       10 * time.Second,
		Prefork:            cfg.GetBool("service.http.prefork"),
		ProxyHeader:        "Cf-Connecting-Ip",
		EnableIPValidation: true,
	})

	app.Use(recover.New(recover.ConfigDefault))

	app.Use(fiberzap.New(fiberzap.Config{
		Logger: logger.L.Desugar(),
		Levels: []zapcore.Level{zapcore.InfoLevel},
		Fields: []string{"latency", "status", "ip", "method", "url"},
	}))

	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Powered-By", "TuruLabs")
		c.Set("X-Developed-By", "Nopaleon Bonaparte")

		return c.Next()
	})

	app.Use(requestid.New(requestid.Config{
		Header: "X-Request-Id",
		Generator: func() string {
			return pkg.GenerateRandomID()
		},
	}))

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return workerpool.Pool.Submit(func() {
				logger.L.Info("fiber server started")
				_ = app.Listen(fmt.Sprintf("%s:%d", cfg.GetString("service.http.host"), cfg.GetInt("service.http.port")))
			})
		},
		OnStop: func(ctx context.Context) error {
			log.Info("fiber server stopped")
			return app.Shutdown()
		},
	})

	return app
}
