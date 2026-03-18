package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Team struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Name       string    `json:"name"`
	Country    string    `json:"country"`
	Budget     int64     `json:"budget"`
	TotalValue int64     `json:"total_value"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Player struct {
	ID          int64     `json:"id"`
	TeamID      int64     `json:"team_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Country     string    `json:"country"`
	Age         int       `json:"age"`
	Position    string    `json:"position"`
	MarketValue int64     `json:"market_value"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TransferListing struct {
	ID          int64     `json:"id"`
	PlayerID    int64     `json:"player_id"`
	AskingPrice int64     `json:"asking_price"`
	ListedAt    time.Time `json:"listed_at"`
}
