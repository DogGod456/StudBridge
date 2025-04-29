package repository

import (
	"context"
	"database/sql"
	"errors"
	"sbChat/internal/models"
	"time"
)

type ChatRepository interface {
	CreateChat(ctx context.Context) (string, error)
	AddParticipant(ctx context.Context, chatID, participantID, senderID string) error
	SendMessage(ctx context.Context, chatID, senderID, text string, isDraft bool) (string, error)
	GetChatMessages(ctx context.Context, chatID string, limit int) ([]models.Message, error)
	GetChatParticipants(ctx context.Context, chatID string) ([]models.Participant, error)
}

type chatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateChat(ctx context.Context) (string, error) {
	var chatID string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO chats DEFAULT VALUES RETURNING id_chat`,
	).Scan(&chatID)
	return chatID, err
}

func (r *chatRepository) AddParticipant(ctx context.Context, chatID, participantID, senderID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO chat_senders (id_chat, id_participant, id_sender) 
		VALUES ($1, $2, $3)`,
		chatID, participantID, senderID,
	)
	return err
}

func (r *chatRepository) SendMessage(ctx context.Context, chatID, senderID, text string, isDraft bool) (string, error) {
	var messageID string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO messages (id_chat, id_sender, message_text, draft) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id_message`,
		chatID, senderID, text, isDraft,
	).Scan(&messageID)
	return messageID, err
}

func (r *chatRepository) GetChatMessages(ctx context.Context, chatID string, limit int) ([]models.Message, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id_message, id_sender, message_text, status, sending_time 
		FROM messages 
		WHERE id_chat = $1 
		ORDER BY sending_time DESC 
		LIMIT $2`,
		chatID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var sendingTime time.Time
		if err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.Text,
			&msg.Status,
			&sendingTime,
		); err != nil {
			return nil, err
		}
		msg.SendingTime = sendingTime
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *chatRepository) GetChatParticipants(ctx context.Context, chatID string) ([]models.Participant, error) {
	// Реализация аналогична GetChatMessages
	return nil, errors.New("not implemented")
}
