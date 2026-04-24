package dto

type RegisterRigRequest struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
}

// TODO: what would be useful here?
type RegisterRigResponse struct{}

type UnregisterRigRequest struct {
	Id string `json:"id"`
}

// TODO: what would be useful here?
type UnregisterRigResponse struct{}

type SendCommandRequest struct {
	ID      string                 `param:"id" json:"id"`
	Command string                 `json:"command"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// TODO: what would be useful here?
type SendCommandResponse struct {
}

type StatusRequest struct {
	ID string `param:"id" json:"id"`
}

type StatusResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
