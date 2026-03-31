package message

type Message struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Retry   int    `json:"retry"`
}
