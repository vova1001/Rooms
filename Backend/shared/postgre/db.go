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
	log.Println("Table 'rooms' ready")

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
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ,
			deleted_at TIMESTAMPTZ
		)
	`)
	if err != nil {
		return fmt.Errorf("create msg: %w", err)
	}
	log.Println("Table 'msg' ready")

		_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS btree_gist`)
	if err != nil {
		return fmt.Errorf("create extension btree_gist: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS events(
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(200) NOT NULL,
			description TEXT,
			price NUMERIC(10,2) NOT NULL DEFAULT 0,
			duration_minutes INT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create events failed: %w", err)
	}
	log.Println("Table 'events' ready")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS event_schedules(
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			date DATE NOT NULL,
			start_time TIME NOT NULL,
			end_time TIME NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create event_schedules failed: %w", err)
	}
	log.Println("Table 'event_schedules' ready")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bookings(
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			start_at TIMESTAMPTZ NOT NULL,
			end_at TIMESTAMPTZ NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			EXCLUDE USING gist (
				event_id WITH =,
				tstzrange(start_at, end_at) WITH &&
			) WHERE (status <> 'cancelled')
		)
	`)
	if err != nil {
		return fmt.Errorf("create bookings failed: %w", err)
	}
	log.Println("Table 'bookings' ready")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications(
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			booking_id UUID REFERENCES bookings(id) ON DELETE CASCADE,
			type VARCHAR(30) NOT NULL,
			text TEXT NOT NULL,
			is_read BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create notifications failed: %w", err)
	}
	log.Println("Table 'notifications' ready")

	


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
