package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sbChat/internal/models"
)

type ChatRepository interface {
	// Управление чатами
	CreateChat(ctx context.Context) (string, error)
	FindChatByParticipants(ctx context.Context, participantIDs ...string) (string, error)
	DeleteChat(ctx context.Context, chatID string) error

	// Управление участниками
	AddParticipant(ctx context.Context, chatID, participantID string) error
	GetChatParticipants(ctx context.Context, chatID string) ([]models.Participant, error)
	IsParticipantInChat(ctx context.Context, chatID, participantID string) (bool, error)
	GetParticipantType(ctx context.Context, participantID string) (string, error)

	// Работа с сообщениями
	SendMessage(ctx context.Context, chatID, senderID, text string, isDraft bool) (string, error)
	GetMessageByID(ctx context.Context, messageID string) (*models.Message, error)
	GetChatMessages(ctx context.Context, chatID string, limit, offset int) ([]models.Message, error)
	UpdateMessageStatus(ctx context.Context, messageID, status string) error
	DeleteMessage(ctx context.Context, messageID string) error

	// Комплексные операции
	CreateChatWithParticipants(ctx context.Context, participantIDs ...string) (string, error)
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

func (r *chatRepository) FindChatByParticipants(ctx context.Context, participantIDs ...string) (string, error) {
	if len(participantIDs) == 0 {
		return "", errors.New("at least one participant is required")
	}

	// Создаем параметры для запроса ($1, $2, ...)
	params := make([]interface{}, len(participantIDs))
	for i, id := range participantIDs {
		params[i] = id
	}

	// Создаем часть запроса с IN условием
	inClause := "("
	for i := 1; i <= len(participantIDs); i++ {
		if i > 1 {
			inClause += ", "
		}
		inClause += fmt.Sprintf("$%d", i)
	}
	inClause += ")"

	query := fmt.Sprintf(`
        SELECT id_chat FROM (
            SELECT id_chat, COUNT(*) as participants_count
            FROM chat_senders
            WHERE id_participant IN %s
            GROUP BY id_chat
        ) AS chats
        WHERE participants_count = %d`,
		inClause, len(participantIDs))

	var chatID string
	err := r.db.QueryRowContext(ctx, query, params...).Scan(&chatID)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return chatID, err
}

func (r *chatRepository) DeleteChat(ctx context.Context, chatID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM chats WHERE id_chat = $1`,
		chatID,
	)
	return err
}

func (r *chatRepository) AddParticipant(ctx context.Context, chatID, participantID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO chat_senders (id_chat, id_participant)
		VALUES ($1, $2)`,
		chatID, participantID,
	)
	return err
}

func (r *chatRepository) GetChatParticipants(ctx context.Context, chatID string) ([]models.Participant, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT p.id_participant, pt.type_name, p.ref_id, p.created_at, p.updated_at
		FROM participants p
		JOIN participant_types pt ON p.id_participant_type = pt.id_participant_type
		JOIN chat_senders cs ON p.id_participant = cs.id_participant
		WHERE cs.id_chat = $1`,
		chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []models.Participant
	for rows.Next() {
		var p models.Participant
		if err := rows.Scan(
			&p.ID,
			&p.ParticipantType,
			&p.RefID,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, nil
}

func (r *chatRepository) IsParticipantInChat(ctx context.Context, chatID, participantID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM chat_senders 
			WHERE id_chat = $1 AND id_participant = $2
		)`,
		chatID, participantID,
	).Scan(&exists)
	return exists, err
}

func (r *chatRepository) GetParticipantType(ctx context.Context, participantID string) (string, error) {
	var participantType string
	err := r.db.QueryRowContext(ctx,
		`SELECT pt.type_name
		FROM participants p
		JOIN participant_types pt ON p.id_participant_type = pt.id_participant_type
		WHERE p.id_participant = $1`,
		participantID,
	).Scan(&participantType)
	return participantType, err
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

func (r *chatRepository) GetMessageByID(ctx context.Context, messageID string) (*models.Message, error) {
	var msg models.Message
	err := r.db.QueryRowContext(ctx,
		`SELECT id_message, id_chat, id_sender, message_text, status, 
				id_parent_message, draft, sending_time, updated_at
		FROM messages
		WHERE id_message = $1`,
		messageID,
	).Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.SenderID,
		&msg.Text,
		&msg.Status,
		&msg.ParentID,
		&msg.IsDraft,
		&msg.SendingTime,
		&msg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *chatRepository) GetChatMessages(ctx context.Context, chatID string, limit, offset int) ([]models.Message, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id_message, id_sender, message_text, status, id_parent_message, 
				draft, sending_time, updated_at
		FROM messages
		WHERE id_chat = $1
		ORDER BY sending_time DESC
		LIMIT $2 OFFSET $3`,
		chatID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.Text,
			&msg.Status,
			&msg.ParentID,
			&msg.IsDraft,
			&msg.SendingTime,
			&msg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *chatRepository) UpdateMessageStatus(ctx context.Context, messageID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE messages 
		SET status = $1, updated_at = NOW()
		WHERE id_message = $2`,
		status, messageID,
	)
	return err
}

func (r *chatRepository) DeleteMessage(ctx context.Context, messageID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM messages WHERE id_message = $1`,
		messageID,
	)
	return err
}

func (r *chatRepository) CreateChatWithParticipants(ctx context.Context, participantIDs ...string) (string, error) {
	if len(participantIDs) < 1 {
		return "", errors.New("at least one participant is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var chatID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO chats DEFAULT VALUES RETURNING id_chat`,
	).Scan(&chatID)
	if err != nil {
		return "", err
	}

	for _, participantID := range participantIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO chat_senders (id_chat, id_participant)
            VALUES ($1, $2)`,
			chatID, participantID,
		)
		if err != nil {
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return chatID, nil
}
