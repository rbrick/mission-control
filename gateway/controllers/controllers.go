package controllers

import (
	"github.com/labstack/echo/v5"
	"github.com/rbrick/mission-control/gateway/hub"
)

func Register(group *echo.Group, h *hub.RigHub) {
	RegisterRigController(group, h)
}
