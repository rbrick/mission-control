package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rbrick/mission-control/gateway/controllers"
)

func main() {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	v1Groups := e.Group("/v1")

	controllers.Register(v1Groups)

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0:8080"
	}

	if err := e.Start(host); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
