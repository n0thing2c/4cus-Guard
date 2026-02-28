package Message

type Message struct {
	Action    string `json:"action"`
	Timestamp int64  `json:"timestamp"`
	URL       string `json:"url"`
}
