package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"bb-project/internal/service"
)

type callbackHandler struct {
	service *service.Callback
}

func newCallbackHandler(task *service.Callback) *callbackHandler {
	return &callbackHandler{task}
}
func (s *callbackHandler) callback(c echo.Context) error {
	req := new(CallbackRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	s.service.Callback(req.ObjectIds)
	return c.JSON(http.StatusOK, "ok") //TODO
}
