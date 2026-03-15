package routes

import (
	"github.com/labstack/echo/v5"
	"github.com/rbrick/mission-control/rig/config"
	"github.com/rbrick/mission-control/rig/dto"
)

type Routes interface {
	GetConfig() *dto.Config

	Register(group *echo.Group)
}

type routesImpl struct {
	config *config.Config
}

func NewRoutes(config *config.Config) Routes {
	return &routesImpl{
		config: config,
	}
}

func (r *routesImpl) Register(group *echo.Group) {
	group.GET("/config", dto.WrapGetter(r.GetConfig))
}

func (r *routesImpl) GetConfig() *dto.Config {
	return &dto.Config{
		ID:          r.config.ID,
		DisplayName: r.config.DisplayName,
		NINA: struct {
			Host string `json:"host"`
		}{
			Host: r.config.NINA.Host,
		},
		Adapter: struct {
			Type string `json:"type"`
		}{
			Type: r.config.Adapter.Type,
		},
	}
}
