package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rbrick/mission-control/gateway/hub"
	"github.com/rbrick/mission-control/gateway/protocol"
)

type RigController struct{ hub *hub.RigHub }

func RegisterRigController(group *echo.Group, h *hub.RigHub) {
	controller := &RigController{hub: h}

	group.GET("/rigs", controller.ListRigs)
	group.GET("/rigs/:id", controller.GetRig)
	group.POST("/rigs/:id/commands", controller.SendCommand)
	group.GET("/commands", controller.ListCommands)
	group.GET("/commands/:id", controller.GetCommand)
	group.GET("/ws/rig", controller.RigWebSocket)

	// Backwards-compatible aliases while the API settles.
	group.GET("/rig/status", controller.ListRigs)
	group.POST("/rig/send", controller.SendCommandLegacy)
}

type sendCommandRequest struct {
	Namespace string                 `json:"namespace"`
	Command   string                 `json:"command"`
	Target    *protocol.Target       `json:"target,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

func (r *RigController) ListRigs(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"rigs": r.hub.ListRigs()})
}

func (r *RigController) GetRig(c *echo.Context) error {
	rig, ok := r.hub.GetRig(c.Param("id"))
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "rig not found"})
	}
	return c.JSON(http.StatusOK, rig)
}

func (r *RigController) SendCommand(c *echo.Context) error {
	return r.sendCommand(c, c.Param("id"))
}

func (r *RigController) ListCommands(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"commands": r.hub.ListCommands()})
}

func (r *RigController) GetCommand(c *echo.Context) error {
	command, ok := r.hub.GetCommand(c.Param("id"))
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "command not found"})
	}
	return c.JSON(http.StatusOK, command)
}

func (r *RigController) SendCommandLegacy(c *echo.Context) error {
	var body struct {
		ID        string                 `json:"id"`
		Namespace string                 `json:"namespace"`
		Command   string                 `json:"command"`
		Params    map[string]interface{} `json:"params,omitempty"`
		Data      map[string]interface{} `json:"data,omitempty"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	data := body.Data
	if data == nil {
		data = body.Params
	}
	return r.dispatch(c, body.ID, body.Namespace, body.Command, nil, data)
}

func (r *RigController) sendCommand(c *echo.Context, rigID string) error {
	var body sendCommandRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	return r.dispatch(c, rigID, body.Namespace, body.Command, body.Target, body.Data)
}

func (r *RigController) dispatch(c *echo.Context, rigID, namespace, command string, target *protocol.Target, data map[string]interface{}) error {
	if rigID == "" || namespace == "" || command == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "rig id, namespace, and command are required"})
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid data"})
	}
	state, err := r.hub.SendCommand(rigID, namespace, command, target, raw)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusAccepted, state)
}

func (r *RigController) RigWebSocket(c *echo.Context) error {
	return r.hub.ServeWS(c.Response(), c.Request())
}
