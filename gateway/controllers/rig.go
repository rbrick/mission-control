package controllers

import (
	"github.com/labstack/echo/v5"
	"github.com/rbrick/mission-control/gateway/dto"
)

/*
What is a rig?

a rig consists of every compontent of an automated telescope.

this would include:
- mount
- focuser
- camera

and might include:
- filter wheel
- dome
- weather station
- power management
- environmental sensors
- safety monitor
- guider

we should include all of these in the info of the rig.

a rig also may or may not support specific actions.

so a rig should tell us what it supports.

for our case, a rig is a node on a network.

the gateway talks with the rig and acts as a control plane for the rig.

sending commands to the rig and receiving telemetry from the rig.

The controller for the rig should be able to tell us what the rig is, what it supports, and how to interact with it.

So there are a few things the gateway needs to know about the rig:
- what components does the rig have?
- what actions does the rig support?
- how do we interact with the rig?
- what is the status of the rig?

as well as be able to register & unregister the rig with the gateway and send commands to the rig.

# To keep a simple protocol i think

- register packet
- send packet
- keep alive packet (telemetry), handled over websockets
*/
type RigController interface {
	Register(*dto.RegisterRigRequest) (*dto.RegisterRigResponse, *dto.Error)
	Unregister(*dto.UnregisterRigRequest) (*dto.UnregisterRigResponse, *dto.Error)
	Send(*dto.SendCommandRequest) (*dto.SendCommandResponse, *dto.Error)
	Status(*dto.StatusRequest) (*dto.StatusResponse, *dto.Error)
}

func RegisterRigController(group *echo.Group) {
	controller := &rigControllerImpl{}

	group.PUT("/rig/register", dto.DTO(controller.Register))
	group.DELETE("/rig/unregister", dto.DTO(controller.Unregister))
	group.POST("/rig/send", dto.DTO(controller.Send))

	group.GET("/rig/status", dto.DTO(controller.Status))
}

type rigControllerImpl struct {
	// to do: implement the rig service
}

func (r *rigControllerImpl) Register(request *dto.RegisterRigRequest) (*dto.RegisterRigResponse, *dto.Error) {

	return &dto.RegisterRigResponse{}, nil
}

func (r *rigControllerImpl) Unregister(request *dto.UnregisterRigRequest) (*dto.UnregisterRigResponse, *dto.Error) {
	return &dto.UnregisterRigResponse{}, nil
}

func (r *rigControllerImpl) Send(request *dto.SendCommandRequest) (*dto.SendCommandResponse, *dto.Error) {
	return &dto.SendCommandResponse{}, nil
}

func (r *rigControllerImpl) Status(request *dto.StatusRequest) (*dto.StatusResponse, *dto.Error) {
	return &dto.StatusResponse{}, nil
}
