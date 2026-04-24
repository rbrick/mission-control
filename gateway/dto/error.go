package dto

type Error struct {
	Message string `json:"message"`
	Status  int    `json:"status,omitempty"`
}

func NewError(message string) *Error {
	return &Error{
		Message: message,
	}
}
