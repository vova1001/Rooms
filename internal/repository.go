package internal

import "database/sql"

type PartRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *PartRepo {
	return &PartRepo{db: db}
}
