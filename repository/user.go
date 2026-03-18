package repository

import (
	"database/sql"

	"github.com/nika/soccer-manager-api/models"
)

func (db *DB) CreateUser(email, passwordHash, fullname string, age int) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO users (email, password_hash, fullname, age, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`, email, passwordHash, fullname, age).Scan(&id)
	return id, err
}

func (db *DB) CreateUserTx(tx *sql.Tx, email, passwordHash, fullname string, age int) (int64, error) {
	var id int64
	err := tx.QueryRow(`
		INSERT INTO users (email, password_hash, fullname, age, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`, email, passwordHash, fullname, age).Scan(&id)
	return id, err
}

func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	var u models.User
	err := db.QueryRow(`
		SELECT id, email, fullname, age, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.FullName, &u.Age, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *DB) GetUserByID(id int64) (*models.User, error) {
	var u models.User
	err := db.QueryRow(`
		SELECT id, email, fullname, age, password_hash, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.FullName, &u.Age, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *DB) ExistsUserByEmail(email string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}
