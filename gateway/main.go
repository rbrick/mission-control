package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rbrick/mission-control/gateway/controllers"
	"github.com/rbrick/mission-control/gateway/hub"
)

func commandTimeout() time.Duration {
	value := os.Getenv("COMMAND_TIMEOUT")
	if value == "" {
		return 30 * time.Second
	}
	timeout, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("invalid COMMAND_TIMEOUT %q, using default: %v", value, err)
		return 30 * time.Second
	}
	return timeout
}

func main() {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	rigHub := hub.New(
		hub.WithToken(os.Getenv("GATEWAY_RIG_TOKEN")),
		hub.WithCommandTimeout(commandTimeout()),
	)
	v1Groups := e.Group("/v1")

	controllers.Register(v1Groups, rigHub)

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0:8080"
	}

	if err := e.Start(host); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
