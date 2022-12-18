package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"bb-project/internal/service"
)

func NewRouter(task *service.Callback) *echo.Echo {
	callbackHandler := newCallbackHandler(task)

	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Logger.Info().
				Str("URI", v.URI).
				Int("status", v.Status).
				Msg("request")
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.POST("/callback", callbackHandler.callback)

	return e
}
