package models

import (
	"time"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

type Message struct {
	ID          string        `json:"id" db:"id_message"`
	ChatID      string        `json:"chat_id" db:"id_chat"`
	SenderID    string        `json:"sender_id" db:"id_sender"`
	Text        string        `json:"text" db:"message_text"`
	Status      MessageStatus `json:"status" db:"status"`
	ParentID    *string       `json:"parent_id,omitempty" db:"id_parent_message"`
	IsDraft     bool          `json:"is_draft" db:"draft"`
	SendingTime time.Time     `json:"sending_time" db:"sending_time"`
	UpdatedAt   *time.Time    `json:"updated_at,omitempty" db:"updated_at"`
}

type Chat struct {
	ID        string     `json:"id" db:"id_chat"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type Participant struct {
	ID              string     `json:"id" db:"id_participant"`
	ParticipantType string     `json:"participant_type" db:"type_name"`
	RefID           string     `json:"ref_id" db:"ref_id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type ChatSender struct {
	ID            string     `json:"-" db:"id_chat_sender"`
	ChatID        string     `json:"-" db:"id_chat"`
	ParticipantID string     `json:"-" db:"id_participant"`
	CreatedAt     time.Time  `json:"-" db:"created_at"`
	UpdatedAt     *time.Time `json:"-" db:"updated_at"`
}

// CreateChatRequest запрос на создание чата
type CreateChatRequest struct {
	ParticipantIDs []string `json:"participant_ids" validate:"required,len=2"`
}

// SendMessageRequest запрос на отправку сообщения
type SendMessageRequest struct {
	ChatID   string  `json:"chat_id" validate:"required,uuid4"`
	Text     string  `json:"text" validate:"required,max=5000"`
	ParentID *string `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
	IsDraft  bool    `json:"is_draft"`
}

// UpdateMessageStatusRequest запрос на обновление статуса сообщения
type UpdateMessageStatusRequest struct {
	MessageID string        `json:"message_id" validate:"required,uuid4"`
	Status    MessageStatus `json:"status" validate:"required,oneof=sent delivered read"`
}
