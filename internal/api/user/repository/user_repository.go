package repository

import (
	"context"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/domain/repositories"

	"gorm.io/gorm"
)

type connectionRepository struct {
	db *gorm.DB
}

func NewConnectionRepository(db *gorm.DB) repositories.ConnectionRepository {
	return &connectionRepository{db: db}
}

func (r *connectionRepository) Create(ctx context.Context, connection *entities.Connection) error {
	return r.db.WithContext(ctx).Create(connection).Error
}

func (r *connectionRepository) GetByID(ctx context.Context, id uint) (*entities.Connection, error) {
	var connection entities.Connection
	err := r.db.WithContext(ctx).
		Preload("Requester").
		Preload("Addressee").
		First(&connection, id).Error
	if err != nil {
		return nil, err
	}
	return &connection, nil
}

func (r *connectionRepository) FindConnection(ctx context.Context, requesterID, addresseeID uint) (*entities.Connection, error) {
	var connection entities.Connection
	err := r.db.WithContext(ctx).
		Where("(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
			requesterID, addresseeID, addresseeID, requesterID).
		First(&connection).Error
	if err != nil {
		return nil, err
	}
	return &connection, nil
}

func (r *connectionRepository) GetUserConnections(ctx context.Context, userID uint, status entities.ConnectionStatus, limit, offset int) ([]*entities.Connection, error) {
	var connections []*entities.Connection
	query := r.db.WithContext(ctx).
		Preload("Requester").
		Preload("Addressee").
		Where("(requester_id = ? OR addressee_id = ?) AND status = ?", userID, userID, status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	err := query.Find(&connections).Error
	return connections, err
}

func (r *connectionRepository) GetConnectionRequests(ctx context.Context, userID uint, limit, offset int) ([]*entities.Connection, error) {
	var connections []*entities.Connection
	err := r.db.WithContext(ctx).
		Preload("Requester").
		Preload("Addressee").
		Where("addressee_id = ? AND status = ?", userID, entities.ConnectionPending).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&connections).Error
	return connections, err
}

func (r *connectionRepository) GetSentRequests(ctx context.Context, userID uint, limit, offset int) ([]*entities.Connection, error) {
	var connections []*entities.Connection
	err := r.db.WithContext(ctx).
		Preload("Requester").
		Preload("Addressee").
		Where("requester_id = ? AND status = ?", userID, entities.ConnectionPending).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&connections).Error
	return connections, err
}

func (r *connectionRepository) Update(ctx context.Context, connection *entities.Connection) error {
	return r.db.WithContext(ctx).Save(connection).Error
}

func (r *connectionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.Connection{}, id).Error
}

func (r *connectionRepository) GetMutualConnections(ctx context.Context, userID1, userID2 uint, limit, offset int) ([]*entities.Connection, error) {
	var connections []*entities.Connection

	subQuery1 := r.db.Select("CASE WHEN requester_id = ? THEN addressee_id ELSE requester_id END as connected_user", userID1).
		Where("(requester_id = ? OR addressee_id = ?) AND status = ?", userID1, userID1, entities.ConnectionAccepted).
		Table("connections")

	subQuery2 := r.db.Select("CASE WHEN requester_id = ? THEN addressee_id ELSE requester_id END as connected_user", userID2).
		Where("(requester_id = ? OR addressee_id = ?) AND status = ?", userID2, userID2, entities.ConnectionAccepted).
		Table("connections")

	err := r.db.WithContext(ctx).
		Preload("Requester").
		Preload("Addressee").
		Where("(requester_id IN (?) OR addressee_id IN (?)) AND status = ?", subQuery1, subQuery1, entities.ConnectionAccepted).
		Where("(requester_id IN (?) OR addressee_id IN (?))", subQuery2, subQuery2).
		Limit(limit).
		Offset(offset).
		Find(&connections).Error

	return connections, err
}
