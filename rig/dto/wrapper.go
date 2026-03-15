package dto

import "github.com/labstack/echo/v5"

func WrapGetter[T any](getter func() *T) echo.HandlerFunc {
	return func(c *echo.Context) error {
		result := getter()
		if result == nil {
			return c.JSON(500, map[string]string{"error": "internal server error"})
		}
		return c.JSON(200, result)
	}
}
