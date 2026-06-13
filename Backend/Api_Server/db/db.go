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
		log.Fatal("error create and open db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error ping db: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	// Включаем расширение pgcrypto для gen_random_uuid()
	_, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
	if err != nil {
		return fmt.Errorf("create extension failed: %w", err)
	}

	// Удаляем таблицы, если они существуют (в правильном порядке: сначала дочерние, потом родительские)
	_, err = db.Exec(`DROP TABLE IF EXISTS rooms CASCADE`)
	if err != nil {
		log.Printf("Warning: drop rooms table: %v", err)
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS users CASCADE`)
	if err != nil {
		log.Printf("Warning: drop users table: %v", err)
	}

	// Создаем таблицу users
	_, err = db.Exec(`CREATE TABLE users(
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        username VARCHAR(50) NOT NULL,
        created_at TIMESTAMPTZ DEFAULT NOW()
    )`)
	if err != nil {
		return fmt.Errorf("create users failed: %w", err)
	}
	log.Println("✓ Table 'users' created")

	// Создаем таблицу rooms с внешним ключом на users
	_, err = db.Exec(`CREATE TABLE rooms(
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name VARCHAR(100) NOT NULL,
        owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        created_at TIMESTAMPTZ DEFAULT NOW()
    )`)
	if err != nil {
		return fmt.Errorf("create rooms failed: %w", err)
	}
	log.Println("✓ Table 'rooms' created")

	return nil
}
