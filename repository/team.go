package repository

import (
	"database/sql"

	"github.com/nika/soccer-manager-api/models"
)

func (db *DB) CreateTeam(userID int64, name, country string, budget, totalValue int64) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO teams (user_id, name, country, budget, total_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`, userID, name, country, budget, totalValue).Scan(&id)
	return id, err
}

func (db *DB) CreateTeamTx(tx *sql.Tx, userID int64, name, country string, budget, totalValue int64) (int64, error) {
	var id int64
	err := tx.QueryRow(`
		INSERT INTO teams (user_id, name, country, budget, total_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`, userID, name, country, budget, totalValue).Scan(&id)
	return id, err
}

func (db *DB) GetTeamByUserID(userID int64) (*models.Team, error) {
	var t models.Team
	err := db.QueryRow(`
		SELECT id, user_id, name, country, budget, total_value, created_at, updated_at
		FROM teams WHERE user_id = $1
	`, userID).Scan(&t.ID, &t.UserID, &t.Name, &t.Country, &t.Budget, &t.TotalValue, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (db *DB) GetTeamByID(id int64) (*models.Team, error) {
	var t models.Team
	err := db.QueryRow(`
		SELECT id, user_id, name, country, budget, total_value, created_at, updated_at
		FROM teams WHERE id = $1
	`, id).Scan(&t.ID, &t.UserID, &t.Name, &t.Country, &t.Budget, &t.TotalValue, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (db *DB) UpdateTeamNameCountry(teamID int64, name, country string) error {
	_, err := db.Exec(`
		UPDATE teams SET name = $1, country = $2, updated_at = NOW() WHERE id = $3
	`, name, country, teamID)
	return err
}

func (db *DB) UpdateTeamBudgetAndValue(teamID int64, budget, totalValue int64) error {
	_, err := db.Exec(`
		UPDATE teams SET budget = $1, total_value = $2, updated_at = NOW() WHERE id = $3
	`, budget, totalValue, teamID)
	return err
}

func (db *DB) UpdateTeamBudgetAndValueTx(tx *sql.Tx, teamID int64, budget, totalValue int64) error {
	_, err := tx.Exec(`
		UPDATE teams SET budget = $1, total_value = $2, updated_at = NOW() WHERE id = $3
	`, budget, totalValue, teamID)
	return err
}
