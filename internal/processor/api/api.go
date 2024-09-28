package api

type Response struct {
	Success     bool   `json:"success"`
	Description string `json:"description"`
	Data        any    `json:"data"`
}
