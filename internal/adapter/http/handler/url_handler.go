package handler

import (
	"URLRotatorGo/internal/adapter/http/dto"
	"URLRotatorGo/internal/core/ports"
	"URLRotatorGo/pkg"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

type URLHandler struct {
	ShortenerService ports.ShortenerService
	cfg              *viper.Viper
}

func NewURLHandler(ShortenerService ports.ShortenerService, cfg *viper.Viper) *URLHandler {
	return &URLHandler{
		ShortenerService: ShortenerService,
		cfg:              cfg,
	}
}

func (h *URLHandler) Index(c *fiber.Ctx) error {
	return c.SendFile("./public/index.html")
}

func (h *URLHandler) RedirectToOriginal(c *fiber.Ctx) error {
	code := c.Params("code")

	redirectUrl, err := h.ShortenerService.GetRedirectURL(c.UserContext(), code)
	if err != nil {
		return c.Status(404).JSON(dto.ApiResponse{
			Error:   true,
			Message: err.Error(),
		})
	}

	return c.Redirect(redirectUrl, 302)
}

func (h *URLHandler) ShortURL(c *fiber.Ctx) error {
	var request dto.RequestShortURL
	var response dto.ApiResponse

	if err := c.BodyParser(&request); err != nil {
		response.Error = true
		response.Message = "url cannot be empty"
		return c.JSON(response)
	}
	if len(request.URL) < 1 {
		response.Error = true
		response.Message = "url cannot be empty"
		return c.JSON(response)
	}
	if len(request.URL) > 100 {
		response.Error = true
		response.Message = "maximum 100 url per request!"
		return c.JSON(response)
	}

	if err := pkg.ValidateRequest(&request); err != nil {
		response.Error = true
		response.Message = err.Error()
		return c.JSON(response)
	}

	result, err := h.ShortenerService.ShortURL(c.UserContext(), request.URL, request.Strategy)
	if err != nil {
		response.Error = true
		response.Message = err.Error()
	}

	response.Data = dto.ResponseShortURL{
		URL:       fmt.Sprintf("%s://%s/%s", h.cfg.GetString("app.scheme"), h.cfg.GetString("app.domain"), result.Code),
		Strategy:  string(result.Strategy),
		CreatedAt: result.CreatedAt,
	}
	return c.JSON(response)
}
