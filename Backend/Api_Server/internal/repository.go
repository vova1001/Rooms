package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"rooms/model"
	m "rooms/model"

	"github.com/google/uuid"
)

type PartRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *PartRepo {
	return &PartRepo{db: db}
}

type UserRepository interface {
	CreateUser(ctx context.Context, username string) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type RoomRepository interface {
	CreateRoom(ctx context.Context, name string, ownerID uuid.UUID) (*model.Room, error)
	GetAllRooms(ctx context.Context) ([]*model.Room, error)
	AddUserToRoom(ctx context.Context, roomID, userID uuid.UUID) error
	GetUsersByRoomID(ctx context.Context, roomID uuid.UUID) ([]*model.User, error)
}

// CreateUser
func (r *PartRepo) CreateUser(ctx context.Context, username string) (*m.User, error) {
	id := uuid.New()
	query := `INSERT INTO users (id, username, created_at) VALUES ($1, $2, NOW()) RETURNING created_at`
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query, id, username).Scan(&createdAt)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("CreateUser cancelled by context: %w", err)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &m.User{
		ID:        id,
		Username:  username,
		CreatedAt: createdAt,
	}, nil
}

// GetUserByID
func (r *PartRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*m.User, error) {
	query := `SELECT id, username, created_at FROM users WHERE id = $1`
	var u m.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Username, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("GetUserByID cancelled: %w", err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id %s: %w", id, err)
	}
	return &u, nil
}

// CreateRoom
func (r *PartRepo) CreateRoom(ctx context.Context, name string, ownerID uuid.UUID) (*m.Room, error) {
	id := uuid.New()
	query := `INSERT INTO rooms (id, name, owner_id, created_at) VALUES ($1, $2, $3, NOW()) RETURNING created_at`
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query, id, name, ownerID).Scan(&createdAt)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("CreateRoom cancelled: %w", err)
		}
		if strings.Contains(err.Error(), "foreign key constraint") {
			return nil, fmt.Errorf("owner_id %s does not exist: %w", ownerID, err)
		}
		return nil, fmt.Errorf("failed to create room: %w", err)
	}
	return &m.Room{
		ID:        id,
		Name:      name,
		OwnerID:   ownerID,
		CreatedAt: createdAt,
	}, nil
}

// GetAllRooms
func (r *PartRepo) GetAllRooms(ctx context.Context) ([]*m.Room, error) {
	query := `SELECT id, name, owner_id, created_at FROM rooms ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("GetAllRooms cancelled: %w", err)
		}
		return nil, fmt.Errorf("failed to query rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*m.Room
	for rows.Next() {
		var rm m.Room
		err := rows.Scan(&rm.ID, &rm.Name, &rm.OwnerID, &rm.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan room error: %w", err)
		}
		rooms = append(rooms, &rm)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return rooms, nil
}

// AddUserToRoom
func (r *PartRepo) AddUserToRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	query := `INSERT INTO room_users (room_id, user_id, joined_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, roomID, userID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("AddUserToRoom cancelled: %w", err)
		}
		return fmt.Errorf("failed to add user %s to room %s: %w", userID, roomID, err)
	}
	return nil
}

// GetUsersByRoomID
func (r *PartRepo) GetUsersByRoomID(ctx context.Context, roomID uuid.UUID) ([]*m.User, error) {
	query := `
        SELECT u.id, u.username, u.created_at 
        FROM users u 
        JOIN room_users ru ON u.id = ru.user_id 
        WHERE ru.room_id = $1
        ORDER BY ru.joined_at
    `
	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("GetUsersByRoomID cancelled: %w", err)
		}
		return nil, fmt.Errorf("failed to query users for room %s: %w", roomID, err)
	}
	defer rows.Close()

	var users []*m.User
	for rows.Next() {
		var u m.User
		err := rows.Scan(&u.ID, &u.Username, &u.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan user error: %w", err)
		}
		users = append(users, &u)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return users, nil
}
