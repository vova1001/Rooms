package postgre

import (
	"database/sql"
	"fmt"
	"log"

	c "backend/shared/configDB"

	_ "github.com/lib/pq"
)

func DBinit(cfgDB *c.ConfigDB) (*sql.DB, error) {
	conectStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfgDB.DBHost, cfgDB.DBPort, cfgDB.DBUser, cfgDB.DBPass, cfgDB.DBName, cfgDB.DBSSLMode)

	db, err := sql.Open("postgres", conectStr)
	if err != nil {
		log.Fatalf("error create and open db: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error ping db: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users(
			id UUID PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			email VARCHAR(254) NOT NULL UNIQUE,
			avatar TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create users failed: %w", err)
	}
	log.Println("Table 'users' ready")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS rooms(
			id UUID PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create rooms failed: %w", err)
	}
	log.Println("✓ Table 'rooms' ready")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS room_users(
			room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			joined_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (room_id, user_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("create room_users failed: %w", err)
	}
	log.Println("Table 'room_users' ready")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS avatars (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			url TEXT NOT NULL UNIQUE,
			is_active BOOLEAN NOT NULL DEFAULT TRUE
		)
	`)
	if err != nil {
		return fmt.Errorf("create avatars failed: %w", err)
	}
	log.Println("Table 'avatars' ready")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages(
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			msg TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)
	`)
	if err != nil {
		return fmt.Errorf("create msg: %w", err)
	}
	log.Println("Table 'msg' ready")

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_msg_room_created
		ON messages(room_id, created_at DESC)
	`)
	if err != nil {
		return fmt.Errorf("create indx_msg_room_created: %w", err)
	}
	log.Println("Table indx_msg ready")

	return nil
}
