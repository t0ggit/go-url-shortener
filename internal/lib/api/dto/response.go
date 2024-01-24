package dto

type Response struct {
	Status string `json:"status"` // ERROR, OK
	Error  string `json:"error,omitempty"`
	Alias  string `json:"alias,omitempty"`
}
