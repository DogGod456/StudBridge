package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sbChat/internal/models"
)

// ParticipantRepository определяет интерфейс для работы с участниками чата
type ParticipantRepository interface {
	// Основные CRUD операции
	CreateParticipant(ctx context.Context, participantType, refID string) (string, error)      // Создание участника
	GetParticipantByID(ctx context.Context, participantID string) (*models.Participant, error) // Получение участника по ID !ПЕРЕДАЕТ ВСЕГО УЧАСТНИКА!
	DeleteParticipant(ctx context.Context, participantID string) error                         // Удаление участника по ID

	// Операции с типами участников
	GetParticipantTypeIDByPartID(ctx context.Context, participantID string) (string, error)     // Получение типа участника по ID участника
	GetParticipantTypeIDByTypeName(ctx context.Context, participantType string) (string, error) // Получение ID типа участника по его названию
	CreateParticipantType(ctx context.Context, participantType string) (string, error)          // Создание нового типа участника

	// Поисковые операции
	FindParticipantByRefAndType(ctx context.Context, refID string, participantType string) (*models.Participant, error) // Поиск участника по ref_id и типу
	GetParticipantByRef(ctx context.Context, refID string) (*models.Participant, error)                                 // Получение участника по ref_id
	GetParticipantIdByRef(ctx context.Context, refID string) (string, error)                                            // Получение ID участника по ref_id
	ExistsParticipantByRef(ctx context.Context, refID string) (bool, error)                                             // Проверка существования участника по ref_id
	ExistsParticipantByIDAndType(ctx context.Context, participantID, participantType string) (bool, error)              // Проверка существования участника по ID и типу
}

// participantRepository реализует ParticipantRepository
type participantRepository struct {
	db *sql.DB
}

// NewParticipantRepository создает новый экземпляр participantRepository
func NewParticipantRepository(db *sql.DB) ParticipantRepository {
	return &participantRepository{db: db}
}

// CreateParticipant создает нового участника в базе данных
func (r *participantRepository) CreateParticipant(ctx context.Context, participantType, refID string) (string, error) {
	var participantID string

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO participants (id_participant_type, ref_id)
		VALUES ($1, $2)
		RETURNING id_participant`,
		participantType, refID,
	).Scan(&participantID)

	if err != nil {
		return "", fmt.Errorf("failed to create participant: %w", err)
	}

	return participantID, nil
}

// GetParticipantByID возвращает участника по его ID
func (r *participantRepository) GetParticipantByID(ctx context.Context, participantID string) (*models.Participant, error) {
	var participant models.Participant

	err := r.db.QueryRowContext(ctx,
		`SELECT p.id_participant, p.id_participant_type, p.ref_id, p.created_at, p.updated_at
		FROM participants p
		WHERE p.id_participant = $1`,
		participantID,
	).Scan(
		&participant.ID,
		&participant.ParticipantType,
		&participant.RefID,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get participant by ID: %w", err)
	}

	return &participant, nil
}

// DeleteParticipant удаляет участника по его ID
func (r *participantRepository) DeleteParticipant(ctx context.Context, participantID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM participants WHERE id_participant = $1`,
		participantID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete participant: %w", err)
	}

	return nil
}

// Операции с типами участников

// GetParticipantTypeIDByPartID возвращает тип участника по его ID
func (r *participantRepository) GetParticipantTypeIDByPartID(ctx context.Context, participantID string) (string, error) {
	var participantTypeID string

	err := r.db.QueryRowContext(ctx,
		`SELECT id_participant_type 
		FROM participants 
		WHERE id_participant = $1`,
		participantID,
	).Scan(&participantTypeID)

	if err != nil {
		return "", fmt.Errorf("failed to get participant type: %w", err)
	}

	return participantTypeID, nil
}

// GetParticipantTypeIDByTypeName возвращает ID типа участника по его названию
func (r *participantRepository) GetParticipantTypeIDByTypeName(ctx context.Context, participantType string) (string, error) {
	var id string

	err := r.db.QueryRowContext(ctx,
		`SELECT id_participant_type::text 
		FROM participant_types 
		WHERE type_name = $1`,
		participantType,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get participant type ID: %w", err)
	}

	return id, nil
}

// CreateParticipantType создает новый тип участника
func (r *participantRepository) CreateParticipantType(ctx context.Context, participantType string) (string, error) {
	var id string

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO participant_types (type_name)
		VALUES ($1)
		RETURNING id_participant_type::text`,
		participantType,
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("failed to create participant type: %w", err)
	}

	return id, nil
}

// Поисковые операции

// FindParticipantByRefAndType ищет участника по ref_id и типу
func (r *participantRepository) FindParticipantByRefAndType(ctx context.Context, refID string, participantType string) (*models.Participant, error) {
	var participant models.Participant

	err := r.db.QueryRowContext(ctx,
		`SELECT p.id_participant, p.id_participant_type, p.ref_id, p.created_at, p.updated_at
		FROM participants p
		JOIN participant_types pt ON p.id_participant_type = pt.id_participant_type
		WHERE p.ref_id = $1 AND pt.type_name = $2`,
		refID, participantType,
	).Scan(
		&participant.ID,
		&participant.ParticipantType,
		&participant.RefID,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find participant by ref and type: %w", err)
	}

	return &participant, nil
}

// GetParticipantByRef возвращает участника по его ref_id
func (r *participantRepository) GetParticipantByRef(ctx context.Context, refID string) (*models.Participant, error) {
	var participant models.Participant

	err := r.db.QueryRowContext(ctx,
		`SELECT p.id_participant, p.id_participant_type, p.ref_id, p.created_at, p.updated_at
		FROM participants p
		WHERE p.ref_id = $1`,
		refID,
	).Scan(
		&participant.ID,
		&participant.ParticipantType,
		&participant.RefID,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get participant by ref: %w", err)
	}

	return &participant, nil
}

// GetParticipantIdByRef возвращает ID участника по его ref_id
func (r *participantRepository) GetParticipantIdByRef(ctx context.Context, refID string) (string, error) {
	var id string

	err := r.db.QueryRowContext(ctx,
		`SELECT p.id_participant
		FROM participants p
		WHERE p.ref_id = $1`,
		refID,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get participant ID by ref: %w", err)
	}

	return id, nil
}

// ExistsParticipantByRef проверяет существование участника с указанным ref_id
func (r *participantRepository) ExistsParticipantByRef(ctx context.Context, refID string) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 
			FROM participants 
			WHERE ref_id = $1
		)`,
		refID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check participant existence: %w", err)
	}

	return exists, nil
}

// ExistsParticipantByIDAndType проверяет существует ли участник с указанным ID и типом
func (r *participantRepository) ExistsParticipantByIDAndType(ctx context.Context, participantID, participantType string) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 
			FROM participants p
			JOIN participant_types pt ON p.id_participant_type = pt.id_participant_type
			WHERE p.id_participant = $1 AND pt.id_participant_type = $2
		)`,
		participantID, participantType,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check participant existence by ID and type: %w", err)
	}

	return exists, nil
}
