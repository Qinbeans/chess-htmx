package static

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// sets the ttl for the static files
const TTL = 3600

func Middleware() func(echo.HandlerFunc) echo.HandlerFunc {
	// add Cache-Control header to static files
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(TTL))
			return next(c)
		}
	}
}
