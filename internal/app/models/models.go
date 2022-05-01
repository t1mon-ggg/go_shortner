package models

type WebData struct {
	Key   string            `json:"key"`
	Short map[string]string `json:"values"`
}

type ClientData struct {
	Cookie string      `json:"cookie"`
	Key    string      `json:"key"`
	Short  []ShortData `json:"values"`
}

type ShortData struct {
	Short string `json:"short"`
	Long  string `json:"long"`
}
