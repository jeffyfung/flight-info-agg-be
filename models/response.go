package model

type Response struct {
	Payload any    `json:"payload"`
	Message string `json:"message,omitempty"`
}
