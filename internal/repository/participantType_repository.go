package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ParticipantTypeRepository определяет интерфейс для работы с типами участников
type ParticipantTypeRepository interface {
	CreateParticipantType(ctx context.Context, typeName string) (string, error)      // Создание типа участника
	DeleteParticipantType(ctx context.Context, typeID string) error                  // Удаление типа участника по ID
	UpdateParticipantType(ctx context.Context, typeID, newTypeName string) error     // Изменение типа участника
	GetParticipantTypeIDByName(ctx context.Context, typeName string) (string, error) // Получение ID типа по названию
	GetParticipantTypeNameByID(ctx context.Context, typeID string) (string, error)   // Получение названия типа по ID
}

// participantTypeRepository реализует ParticipantTypeRepository
type participantTypeRepository struct {
	db *sql.DB
}

// NewParticipantTypeRepository создает новый экземпляр participantTypeRepository
func NewParticipantTypeRepository(db *sql.DB) ParticipantTypeRepository {
	return &participantTypeRepository{db: db}
}

// CreateParticipantType создает новый тип участника в базе данных
func (r *participantTypeRepository) CreateParticipantType(ctx context.Context, typeName string) (string, error) {
	var typeID string

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO participant_types (type_name)
		VALUES ($1)
		RETURNING id_participant_type::text`,
		typeName,
	).Scan(&typeID)

	if err != nil {
		return "", fmt.Errorf("failed to create participant type: %w", err)
	}

	return typeID, nil
}

// DeleteParticipantType удаляет тип участника по его ID
func (r *participantTypeRepository) DeleteParticipantType(ctx context.Context, typeID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM participant_types WHERE id_participant_type = $1`,
		typeID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete participant type: %w", err)
	}

	return nil
}

// UpdateParticipantType обновляет название типа участника
func (r *participantTypeRepository) UpdateParticipantType(ctx context.Context, typeID, newTypeName string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE participant_types 
		SET type_name = $1, updated_at = NOW()
		WHERE id_participant_type = $2`,
		newTypeName, typeID,
	)

	if err != nil {
		return fmt.Errorf("failed to update participant type: %w", err)
	}

	return nil
}

// GetParticipantTypeIDByName возвращает ID типа участника по его названию
func (r *participantTypeRepository) GetParticipantTypeIDByName(ctx context.Context, typeName string) (string, error) {
	var id string

	err := r.db.QueryRowContext(ctx,
		`SELECT id_participant_type::text 
		FROM participant_types 
		WHERE type_name = $1`,
		typeName,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get participant type ID: %w", err)
	}

	return id, nil
}

// GetParticipantTypeNameByID возвращает название типа участника по его ID
func (r *participantTypeRepository) GetParticipantTypeNameByID(ctx context.Context, typeID string) (string, error) {
	var typeName string

	err := r.db.QueryRowContext(ctx,
		`SELECT type_name 
		FROM participant_types 
		WHERE id_participant_type = $1`,
		typeID,
	).Scan(&typeName)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get participant type name: %w", err)
	}

	return typeName, nil
}
