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
	if age < 0 || age > 90 {
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
