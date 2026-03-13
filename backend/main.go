package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0:8080"
	}

	if err := e.Start(host); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
