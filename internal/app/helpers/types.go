package helpers

type Data map[string]WebData

type WebData struct {
	Key   string            `json:"key"`
	Short map[string]string `json:"values"`
}
