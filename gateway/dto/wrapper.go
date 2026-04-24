package dto

import "github.com/labstack/echo/v5"

func Wrap[T any](getter func(*echo.Context) *T) echo.HandlerFunc {
	return func(c *echo.Context) error {
		result := getter(c)
		if result == nil {
			return c.JSON(500, NewError("internal server error"))
		}
		return c.JSON(200, result)
	}
}

// Helper function for creating DTO handlers. It takes a getter function that retrieves the data to be returned in the response, and a variadic list of extractor functions that can be used to extract data from the request context and populate the input data structure.
func DTO[D, T any](getter func(*D) (*T, *Error)) echo.HandlerFunc {
	return func(c *echo.Context) error {
		var data D

		if err := c.Bind(&data); err != nil {
			return c.JSON(400, NewError("invalid request body"))
		}

		result, err := getter(&data)

		if err != nil {
			if err.Status == 0 {
				err.Status = 500
			}
			return c.JSON(err.Status, err)
		}

		return c.JSON(200, result)
	}
}
