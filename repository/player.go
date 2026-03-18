package repository

import (
	"database/sql"

	"github.com/nika/soccer-manager-api/models"
)

func (db *DB) CreatePlayer(teamID int64, firstName, lastName, country string, age int, position string, marketValue int64) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO players (team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`, teamID, firstName, lastName, country, age, position, marketValue).Scan(&id)
	return id, err
}

func (db *DB) CreatePlayersTx(tx *sql.Tx, teamID int64, players []models.Player) error {
	for _, p := range players {
		_, err := tx.Exec(`
			INSERT INTO players (team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		`, teamID, p.FirstName, p.LastName, p.Country, p.Age, p.Position, p.MarketValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) GetPlayersByTeamID(teamID int64) ([]models.Player, error) {
	rows, err := db.Query(`
		SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at
		FROM players WHERE team_id = $1 ORDER BY id
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Player
	for rows.Next() {
		var p models.Player
		if err := rows.Scan(&p.ID, &p.TeamID, &p.FirstName, &p.LastName, &p.Country, &p.Age, &p.Position, &p.MarketValue, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (db *DB) GetPlayerByID(id int64) (*models.Player, error) {
	var p models.Player
	err := db.QueryRow(`
		SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at
		FROM players WHERE id = $1
	`, id).Scan(&p.ID, &p.TeamID, &p.FirstName, &p.LastName, &p.Country, &p.Age, &p.Position, &p.MarketValue, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *DB) UpdatePlayerDetails(id int64, firstName, lastName, country string) error {
	_, err := db.Exec(`
		UPDATE players SET first_name = $1, last_name = $2, country = $3, updated_at = NOW() WHERE id = $4
	`, firstName, lastName, country, id)
	return err
}

func (db *DB) UpdatePlayerTeamAndValue(id int64, teamID int64, marketValue int64) error {
	_, err := db.Exec(`
		UPDATE players SET team_id = $1, market_value = $2, updated_at = NOW() WHERE id = $3
	`, teamID, marketValue, id)
	return err
}

func (db *DB) UpdatePlayerTeamAndValueTx(tx *sql.Tx, id int64, teamID int64, marketValue int64) error {
	_, err := tx.Exec(`
		UPDATE players SET team_id = $1, market_value = $2, updated_at = NOW() WHERE id = $3
	`, teamID, marketValue, id)
	return err
}
