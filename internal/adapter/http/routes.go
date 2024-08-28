package http

import (
	"URLRotatorGo/internal/adapter/http/dto"
	"URLRotatorGo/internal/adapter/http/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"time"
)

type Router struct {
	app        *fiber.App
	urlHandler *handler.URLHandler
}

func NewRouter(
	app *fiber.App,
	urlHandler *handler.URLHandler,
) *Router {
	return &Router{
		app:        app,
		urlHandler: urlHandler,
	}
}

func (r *Router) SetupRoutes() {
	route := r.app.Group("")

	route.Get("/", etag.New(etag.Config{
		Weak: true,
	}), r.urlHandler.Index)
	route.Use("/api/shorten", limiter.New(limiter.Config{
		Max:        15,
		Expiration: time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(dto.ApiResponse{
				Error:   true,
				Message: "too many requests",
			})
		},
		LimiterMiddleware: limiter.SlidingWindow{},
	}))
	route.Post("/api/shorten", r.urlHandler.ShortURL)
	route.Get("/:code", r.urlHandler.RedirectToOriginal)
}
