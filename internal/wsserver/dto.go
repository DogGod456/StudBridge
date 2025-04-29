package wsserver

type WsRequest struct {
	ChatID   string `json:"chat_id"`
	SenderID string `json:"sender_id"`
	Text     string `json:"text"`
}

type WsResponse struct {
	MessageID string `json:"message_id"`
	ChatID    string `json:"chat_id"`
	SenderID  string `json:"sender_id"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
	Status    string `json:"status"`
}
