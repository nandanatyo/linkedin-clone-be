// internal/domain/repositories/connection_repository.go
package repositories

import (
	"context"
	"linked-clone/internal/domain/entities"
)

type ConnectionRepository interface {
	Create(ctx context.Context, connection *entities.Connection) error
	GetByID(ctx context.Context, id uint) (*entities.Connection, error)
	FindConnection(ctx context.Context, requesterID, addresseeID uint) (*entities.Connection, error)
	GetUserConnections(ctx context.Context, userID uint, status entities.ConnectionStatus, limit, offset int) ([]*entities.Connection, error)
	GetConnectionRequests(ctx context.Context, userID uint, limit, offset int) ([]*entities.Connection, error)
	GetSentRequests(ctx context.Context, userID uint, limit, offset int) ([]*entities.Connection, error)
	Update(ctx context.Context, connection *entities.Connection) error
	Delete(ctx context.Context, id uint) error
	GetMutualConnections(ctx context.Context, userID1, userID2 uint, limit, offset int) ([]*entities.Connection, error)
}
