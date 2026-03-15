package dto

type Config struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`

	NINA struct {
		Host string `json:"host"`
	} `json:"nina,omitempty"`

	Adapter struct {
		Type string `json:"type"`
	} `json:"adapter"`
}
