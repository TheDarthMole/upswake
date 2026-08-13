package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v5"
)

func setSlogAttributes(c *echo.Context, attributes []slog.Attr) {
	c.Set("request_logger_values", attributes)
}
