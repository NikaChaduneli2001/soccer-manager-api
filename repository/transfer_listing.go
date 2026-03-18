package repository

import (
	"database/sql"

	"github.com/nika/soccer-manager-api/models"
)

func (db *DB) CreateTransferListing(playerID int64, askingPrice int64) error {
	_, err := db.Exec(`
		INSERT INTO transfer_listings (player_id, asking_price, listed_at)
		VALUES ($1, $2, NOW())
	`, playerID, askingPrice)
	return err
}

func (db *DB) GetTransferListingByID(id int64) (*models.TransferListing, error) {
	var l models.TransferListing
	err := db.QueryRow(`
		SELECT id, player_id, asking_price, listed_at
		FROM transfer_listings WHERE id = $1
	`, id).Scan(&l.ID, &l.PlayerID, &l.AskingPrice, &l.ListedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (db *DB) GetTransferListingByPlayerID(playerID int64) (*models.TransferListing, error) {
	var l models.TransferListing
	err := db.QueryRow(`
		SELECT id, player_id, asking_price, listed_at
		FROM transfer_listings WHERE player_id = $1
	`, playerID).Scan(&l.ID, &l.PlayerID, &l.AskingPrice, &l.ListedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (db *DB) DeleteTransferListingByPlayerID(playerID int64) error {
	_, err := db.Exec(`DELETE FROM transfer_listings WHERE player_id = $1`, playerID)
	return err
}

func (db *DB) DeleteTransferListingByPlayerIDTx(tx *sql.Tx, playerID int64) error {
	_, err := tx.Exec(`DELETE FROM transfer_listings WHERE player_id = $1`, playerID)
	return err
}

// TransferListingWithPlayer holds listing + player for market list.
type TransferListingWithPlayer struct {
	Listing    models.TransferListing
	Player     models.Player
	SellerTeam string // team name for display
}

func (db *DB) ListTransferListings() ([]TransferListingWithPlayer, error) {
	rows, err := db.Query(`
		SELECT t.id, t.player_id, t.asking_price, t.listed_at,
		       p.id, p.team_id, p.first_name, p.last_name, p.country, p.age, p.position, p.market_value, p.created_at, p.updated_at,
		       COALESCE(te.name, '') as seller_team_name
		FROM transfer_listings t
		JOIN players p ON p.id = t.player_id
		JOIN teams te ON te.id = p.team_id
		ORDER BY t.listed_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []TransferListingWithPlayer
	for rows.Next() {
		var item TransferListingWithPlayer
		var p models.Player
		err := rows.Scan(
			&item.Listing.ID, &item.Listing.PlayerID, &item.Listing.AskingPrice, &item.Listing.ListedAt,
			&p.ID, &p.TeamID, &p.FirstName, &p.LastName, &p.Country, &p.Age, &p.Position, &p.MarketValue, &p.CreatedAt, &p.UpdatedAt,
			&item.SellerTeam,
		)
		if err != nil {
			return nil, err
		}
		item.Player = p
		list = append(list, item)
	}
	return list, rows.Err()
}
