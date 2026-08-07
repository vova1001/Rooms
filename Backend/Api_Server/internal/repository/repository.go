package repository

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
	"github.com/redis/go-redis/v9"
)

type PartRepo struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewRepo(db *sql.DB, rdb *redis.Client) *PartRepo {
	return &PartRepo{db: db, rdb: rdb}
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
func (r *PartRepo) CreateUser(ctx context.Context, username, email, avatar string) (*m.User, error) {
	id := uuid.New()
	query := `INSERT INTO users (id, username, email, avatar, created_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING created_at`
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query, id, username, email, avatar).Scan(&createdAt)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("CreateUser cancelled by context: %w", err)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &m.User{
		ID:        id,
		Username:  username,
		Email:     email,
		Avatar:    avatar,
		CreatedAt: createdAt,
	}, nil
}

// GetUserByID
func (r *PartRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*m.User, error) {
	query := `SELECT id, username, created_at, avatar, email FROM users WHERE id = $1`
	var u m.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Username, &u.CreatedAt, &u.Avatar, &u.Email)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("GetUserByID cancelled: %w", err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", id)
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
		RoomName:  name,
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
		var ru []m.RoomUser

		err := rows.Scan(&rm.ID, &rm.RoomName, &rm.OwnerID, &rm.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan room error: %w", err)
		}
		userIDs, err := r.rdb.SMembers(ctx, "room:"+rm.ID.String()+":users").Result()
		if err != nil {
			return nil, fmt.Errorf("err redis result userId in room: %w", err)
		}
		pipe := r.rdb.Pipeline()
		cmds := make(map[string]*redis.MapStringStringCmd)

		for _, userID := range userIDs {
			cmds[userID] = pipe.HGetAll(ctx, "user:"+userID)
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("redis pipeline error: %w", err)
		}

		for _, cmd := range cmds {
			userMap, err := cmd.Result()
			if err != nil {
				return nil, fmt.Errorf("redis cmd error: %w", err)
			}

			id, err := uuid.Parse(userMap["id"])
			if err != nil {
				return nil, fmt.Errorf("invalid uuid: %w", err)
			}
			createdAtUser, err := time.Parse(time.RFC3339Nano, userMap["created_at"])
			if err != nil {
				return nil, fmt.Errorf("parse user created_at error: %w", err)
			}

			ru = append(ru, m.RoomUser{
				RoomID: rm.ID,
				UserInfo: m.User{
					ID:        id,
					Username:  userMap["user_name"],
					CreatedAt: createdAtUser,
					Avatar:    userMap["avatar"],
				},
			})
		}
		rm.RoomUsers = ru
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
        SELECT u.id, u.username, u.created_at, u.avatar 
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
		err := rows.Scan(&u.ID, &u.Username, &u.CreatedAt, &u.Avatar)
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

func (r *PartRepo) EmailCheck(ctx context.Context, email string) (bool, error) {
	query := "SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)"
	var exists bool

	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			return false, fmt.Errorf("emailCheck cancelled by ctx: %w", err)
		}
		return false, fmt.Errorf("failed emailCheck: %w", err)
	}

	return exists, nil
}

func (r *PartRepo) FindUserByEmail(ctx context.Context, email string) (*m.User, error) {
	var user m.User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, email, created_at, avatar FROM users WHERE email =$1 LIMIT 1", email).Scan(
		&user.ID,
		&user.Username,
		&user.CreatedAt,
		&user.Avatar,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &user, nil
}

func (r *PartRepo) GetAvatars(ctx context.Context) ([]Avatars, error) {

	avatars := make([]Avatars, 0)

	query := "SELECT id, url FROM avatars WHERE is_active = $1"

	rows, err := r.db.QueryContext(ctx, query, 1)
	if err != nil {
		return nil, fmt.Errorf("err create rows: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var avatar Avatars
		if err := rows.Scan(&avatar.id, &avatar.url); err != nil {
			return nil, fmt.Errorf("err scan avatar: %w", err)
		}
		avatars = append(avatars, avatar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return avatars, nil
}
