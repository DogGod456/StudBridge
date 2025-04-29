package models

import "time"

type Message struct {
	ID          string    `json:"id"`
	ChatID      string    `json:"chat_id"`
	SenderID    string    `json:"sender_id"`
	Text        string    `json:"text"`
	Status      string    `json:"status"`
	ParentID    string    `json:"parent_id,omitempty"`
	IsDraft     bool      `json:"is_draft"`
	SendingTime time.Time `json:"sending_time"`
}

type Chat struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type Participant struct {
	ID              string    `json:"id"`
	ParticipantType string    `json:"participant_type"`
	RefID           string    `json:"ref_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}

type ChatSender struct {
	ID            string    `json:"id"`
	ChatID        string    `json:"chat_id"`
	SenderID      string    `json:"sender_id"`
	ParticipantID string    `json:"participant_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}
