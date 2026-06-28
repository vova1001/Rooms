package db

import (
	"database/sql"
	"fmt"
	"log"

	c "rooms/config"

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
	_, err := db.Exec(`DROP TABLE IF EXISTS room_users CASCADE`)
	if err != nil {
		log.Printf("Warning: drop room_users table: %v", err)
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS rooms CASCADE`)
	if err != nil {
		log.Printf("Warning: drop rooms table: %v", err)
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS users CASCADE`)
	if err != nil {
		log.Printf("Warning: drop users table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE users(
        id UUID PRIMARY KEY,
        username VARCHAR(50) NOT NULL,
        created_at TIMESTAMPTZ DEFAULT NOW()
    )`)
	if err != nil {
		return fmt.Errorf("create users failed: %w", err)
	}
	log.Println("✓ Table 'users' created")

	_, err = db.Exec(`CREATE TABLE rooms(
        id UUID PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        created_at TIMESTAMPTZ DEFAULT NOW()
    )`)
	if err != nil {
		return fmt.Errorf("create rooms failed: %w", err)
	}
	log.Println("✓ Table 'rooms' created")

	_, err = db.Exec(`CREATE TABLE room_users(
        room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        joined_at TIMESTAMPTZ DEFAULT NOW(),
        PRIMARY KEY (room_id, user_id)
    )`)
	if err != nil {
		return fmt.Errorf("create room_users failed: %w", err)
	}
	log.Println("✓ Table 'room_users' created")

	return nil
}
