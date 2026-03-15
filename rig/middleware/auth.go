package middleware

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func AuthMiddleware(token string) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("X-Auth-Token")
			if authHeader != token {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Unauthorized",
				})
			}
			return next(c)
		}
	}
}
