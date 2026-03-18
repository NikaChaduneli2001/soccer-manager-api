package service

import (
	"database/sql"
	"errors"
	"math/rand"
	"regexp"
	"strconv"

	"github.com/nika/soccer-manager-api/models"
	"github.com/nika/soccer-manager-api/pkg/auth"
	"github.com/nika/soccer-manager-api/repository"
)

const (
	InitialBudget      = 5_000_000 // $5,000,000 per team
	InitialPlayerValue = 1_000_000 // $1,000,000 per player
	Goalkeepers        = 3
	Defenders          = 6
	Midfielders        = 6
	Attackers          = 5
	TotalPlayers       = 20
)

var (
	ErrEmailInvalid       = errors.New("invalid email")
	ErrEmailExists        = errors.New("email already registered")
	ErrPasswordShort      = errors.New("password too short")
	ErrAgeInvalid         = errors.New("age must be between 0 and 150")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTeamNotFound       = errors.New("team not found")
	ErrPlayerNotFound     = errors.New("player not found")
	ErrNotYourPlayer      = errors.New("player does not belong to your team")
	ErrAlreadyListed      = errors.New("player is already on transfer list")
	ErrInsufficientBudget = errors.New("insufficient budget")
	ErrListingNotFound    = errors.New("listing not found")
	ErrCannotBuyOwnPlayer = errors.New("cannot buy your own player")
	ErrTeamAlreadyExists  = errors.New("user already has a team")
	ErrInvalidPosition    = errors.New("invalid position: must be goalkeeper, defender, midfielder, or attacker")
	ErrUserNotFound       = errors.New("user not found")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type Service struct {
	Repo           *repository.DB
	JWTSecret      string
	JWTExpireHours int
}

func NewService(repo *repository.DB, jwtSecret string, jwtExpireHours int) *Service {
	return &Service{Repo: repo, JWTSecret: jwtSecret, JWTExpireHours: jwtExpireHours}
}

// Signup creates a user, a team, and 20 players (3 GK, 6 DEF, 6 MID, 5 ATT), each value $1M, budget $5M.
// All DB writes run in a single transaction; on any failure everything is rolled back.
func (s *Service) Signup(email, password, fullname string, age int) (*models.AuthResponse, error) {
	if !emailRegex.MatchString(email) {
		return nil, ErrEmailInvalid
	}
	if len(password) < 6 {
		return nil, ErrPasswordShort
	}
	if age <= 0 || age > 90 {
		return nil, ErrAgeInvalid
	}
	exists, err := s.Repo.ExistsUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	tx, err := s.Repo.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	userID, err := s.Repo.CreateUserTx(tx, email, hash, fullname, age)
	if err != nil {
		return nil, err
	}
	totalValue := int64(TotalPlayers) * InitialPlayerValue
	teamID, err := s.Repo.CreateTeamTx(tx, userID, "", "", int64(InitialBudget), totalValue)
	if err != nil {
		return nil, err
	}
	players := generateInitialSquad(teamID)
	if err := s.Repo.CreatePlayersTx(tx, teamID, players); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	token, err := s.generateToken(userID)
	if err != nil {
		return nil, err
	}
	return &models.AuthResponse{UserID: userID, Email: email, FullName: fullname, Age: age, Token: token}, nil
}

func (s *Service) Login(email, password string) (*models.AuthResponse, error) {
	u, err := s.Repo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !auth.CheckPassword(u.Password, password) {
		return nil, ErrInvalidCredentials
	}
	token, err := s.generateToken(u.ID)
	if err != nil {
		return nil, err
	}
	return &models.AuthResponse{UserID: u.ID, Email: u.Email, FullName: u.FullName, Age: u.Age, Token: token}, nil
}

// CreateUser creates a user account only (no team, no players). For use with separate team creation.
func (s *Service) CreateUser(email, password, fullname string, age int) (*models.User, error) {
	if !emailRegex.MatchString(email) {
		return nil, ErrEmailInvalid
	}
	if len(password) < 6 {
		return nil, ErrPasswordShort
	}
	if age < 0 || age > 150 {
		return nil, ErrAgeInvalid
	}
	exists, err := s.Repo.ExistsUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	userID, err := s.Repo.CreateUser(email, hash, fullname, age)
	if err != nil {
		return nil, err
	}
	return s.Repo.GetUserByID(userID)
}

func (s *Service) generateToken(userID int64) (string, error) {
	return auth.NewToken(userID, s.JWTSecret, s.JWTExpireHours)
}

func generateInitialSquad(teamID int64) []models.Player {
	var players []models.Player
	positions := []struct {
		pos string
		n   int
	}{
		{"goalkeeper", Goalkeepers},
		{"defender", Defenders},
		{"midfielder", Midfielders},
		{"attacker", Attackers},
	}
	idx := 1
	for _, p := range positions {
		for i := 0; i < p.n; i++ {
			age := 18 + rand.Intn(23) // 18..40
			players = append(players, models.Player{
				TeamID:      teamID,
				FirstName:   "Player",
				LastName:    strconv.Itoa(idx),
				Country:     "",
				Age:         age,
				Position:    p.pos,
				MarketValue: InitialPlayerValue,
			})
			idx++
		}
	}
	return players
}

// GetTeam returns the authenticated user's team (total_value = sum of player values).
func (s *Service) GetTeam(userID int64) (*models.Team, error) {
	team, err := s.Repo.GetTeamByUserID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	players, err := s.Repo.GetPlayersByTeamID(team.ID)
	if err != nil {
		return nil, err
	}
	var totalValue int64
	for _, p := range players {
		totalValue += p.MarketValue
	}
	team.TotalValue = totalValue
	return team, nil
}

func (s *Service) CreateTeam(userID int64, name, country string, budget int64) (*models.Team, error) {
	_, err := s.Repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	_, err = s.Repo.GetTeamByUserID(userID)
	if err == nil {
		return nil, ErrTeamAlreadyExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if budget <= 0 {
		budget = int64(InitialBudget)
	}
	tx, err := s.Repo.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	totalValue := int64(TotalPlayers) * InitialPlayerValue
	teamID, err := s.Repo.CreateTeamTx(tx, userID, name, country, budget, totalValue)
	if err != nil {
		return nil, err
	}
	players := generateInitialSquad(teamID)
	if err := s.Repo.CreatePlayersTx(tx, teamID, players); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	team, err := s.Repo.GetTeamByID(teamID)
	if err != nil {
		return nil, err
	}
	team.TotalValue = totalValue
	return team, nil
}

func (s *Service) UpdateTeam(userID int64, name, country string) error {
	team, err := s.Repo.GetTeamByUserID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTeamNotFound
		}
		return err
	}
	return s.Repo.UpdateTeamNameCountry(team.ID, name, country)
}

// GetTeamPlayers returns all players for the authenticated user's team.
func (s *Service) GetTeamPlayers(userID int64) ([]models.Player, error) {
	team, err := s.Repo.GetTeamByUserID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	return s.Repo.GetPlayersByTeamID(team.ID)
}

var validPositions = map[string]bool{
	"goalkeeper": true, "defender": true, "midfielder": true, "attacker": true,
}

func (s *Service) CreatePlayer(userID int64, firstName, lastName, country, position string) (*models.Player, error) {
	team, err := s.Repo.GetTeamByUserID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	if !validPositions[position] {
		return nil, ErrInvalidPosition
	}
	age := 18 + rand.Intn(23) // 18..40
	playerID, err := s.Repo.CreatePlayer(team.ID, firstName, lastName, country, age, position, InitialPlayerValue)
	if err != nil {
		return nil, err
	}
	return s.Repo.GetPlayerByID(playerID)
}

func (s *Service) UpdatePlayer(userID int64, playerID int64, firstName, lastName, country string) error {
	team, err := s.Repo.GetTeamByUserID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTeamNotFound
		}
		return err
	}
	player, err := s.Repo.GetPlayerByID(playerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPlayerNotFound
		}
		return err
	}
	if player.TeamID != team.ID {
		return ErrNotYourPlayer
	}
	return s.Repo.UpdatePlayerDetails(playerID, firstName, lastName, country)
}

func (s *Service) ListPlayerOnTransfer(userID int64, playerID int64, askingPrice int64) error {
	team, err := s.Repo.GetTeamByUserID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTeamNotFound
		}
		return err
	}
	player, err := s.Repo.GetPlayerByID(playerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPlayerNotFound
		}
		return err
	}
	if player.TeamID != team.ID {
		return ErrNotYourPlayer
	}
	if askingPrice <= 0 {
		return errors.New("asking price must be positive")
	}
	_, err = s.Repo.GetTransferListingByPlayerID(playerID)
	if err == nil {
		return ErrAlreadyListed
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return s.Repo.CreateTransferListing(playerID, askingPrice)
}

// GetTransferList returns all players on the transfer market with asking price.
func (s *Service) GetTransferList() ([]models.TransferMarketItem, error) {
	items, err := s.Repo.ListTransferListings()
	if err != nil {
		return nil, err
	}
	out := make([]models.TransferMarketItem, len(items))
	for i := range items {
		out[i] = models.TransferMarketItem{
			ListingID:   items[i].Listing.ID,
			PlayerID:    items[i].Player.ID,
			AskingPrice: items[i].Listing.AskingPrice,
			FirstName:   items[i].Player.FirstName,
			LastName:    items[i].Player.LastName,
			Country:     items[i].Player.Country,
			Age:         items[i].Player.Age,
			Position:    items[i].Player.Position,
			MarketValue: items[i].Player.MarketValue,
			ListedAt:    items[i].Listing.ListedAt,
		}
	}
	return out, nil
}

func (s *Service) BuyPlayer(buyerUserID int64, listingID int64) error {
	listing, err := s.Repo.GetTransferListingByID(listingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrListingNotFound
		}
		return err
	}
	player, err := s.Repo.GetPlayerByID(listing.PlayerID)
	if err != nil {
		return err
	}
	sellerTeam, err := s.Repo.GetTeamByID(player.TeamID)
	if err != nil {
		return err
	}
	buyerTeam, err := s.Repo.GetTeamByUserID(buyerUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTeamNotFound
		}
		return err
	}
	if sellerTeam.ID == buyerTeam.ID {
		return ErrCannotBuyOwnPlayer
	}
	if buyerTeam.Budget < listing.AskingPrice {
		return ErrInsufficientBudget
	}
	// 10% to 100% random increase on transfer (e.g. 1.1 to 2.0 multiplier)
	increase := 1.0 + 0.1*float64(rand.Intn(10)+1) // 1.1 .. 2.0
	newValue := int64(float64(player.MarketValue) * increase)
	if newValue < player.MarketValue {
		newValue = player.MarketValue
	}
	tx, err := s.Repo.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := s.Repo.UpdatePlayerTeamAndValueTx(tx, player.ID, buyerTeam.ID, newValue); err != nil {
		return err
	}
	if err := s.Repo.UpdateTeamBudgetAndValueTx(tx, sellerTeam.ID, sellerTeam.Budget+listing.AskingPrice, sellerTeam.TotalValue-player.MarketValue); err != nil {
		return err
	}
	if err := s.Repo.UpdateTeamBudgetAndValueTx(tx, buyerTeam.ID, buyerTeam.Budget-listing.AskingPrice, buyerTeam.TotalValue+newValue); err != nil {
		return err
	}
	if err := s.Repo.DeleteTransferListingByPlayerIDTx(tx, player.ID); err != nil {
		return err
	}
	return tx.Commit()
}
