package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
	"github.com/rbrick/mission-control/rig/config"
	rigMiddleware "github.com/rbrick/mission-control/rig/middleware"
	"github.com/rbrick/mission-control/rig/routes"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var configPath string
	var listenAddr string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the rig API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			return runServe(ctx, listenAddr, configPath)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to a rig config YAML file")
	cmd.Flags().StringVar(&listenAddr, "listen", defaultListenAddr(), "Address for the HTTP server")

	return cmd
}

func defaultListenAddr() string {
	if host := os.Getenv("HOST"); host != "" {
		return host
	}

	return "0.0.0.0:8081"
}

func runServe(ctx context.Context, listenAddr, configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	e := echo.New()
	e.Use(echoMiddleware.RequestLogger())
	e.Use(echoMiddleware.Recover())

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	api := e.Group("/api")
	if cfg.Auth.Token != "" {
		api.Use(rigMiddleware.AuthMiddleware(cfg.Auth.Token))
	}

	routes.NewRoutes(cfg).Register(api)

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           e,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	log.Printf("rig %s listening on %s", cfg.ID, listenAddr)

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown rig server: %w", err)
		}

		return nil
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	}
}
