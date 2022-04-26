package models

type WebData struct {
	Key   string            `json:"key"`
	Short map[string]string `json:"values"`
}
