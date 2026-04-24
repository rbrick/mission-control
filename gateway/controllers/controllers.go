package controllers

import "github.com/labstack/echo/v5"

func Register(group *echo.Group) {
	RegisterRigController(group)
}
