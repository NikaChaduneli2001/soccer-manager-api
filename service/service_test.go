package service

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nika/soccer-manager-api/pkg/auth"
	"github.com/nika/soccer-manager-api/repository"
)

func newMockService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	repo := &repository.DB{DB: db}
	svc := NewService(repo, "test-jwt-secret", 24)
	return svc, mock
}

func TestNewService(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	repo := &repository.DB{DB: db}
	svc := NewService(repo, "secret", 24)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.Repo != repo || svc.JWTSecret != "secret" || svc.JWTExpireHours != 24 {
		t.Errorf("NewService: wrong fields")
	}
}

// --- Signup ---

func TestSignup_Validation(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	tests := []struct {
		name     string
		email    string
		password string
		fullname string
		age      int
		wantErr  error
	}{
		{"invalid email", "not-an-email", "password123", "Full", 25, ErrEmailInvalid},
		{"short password", "a@b.com", "short", "Full", 25, ErrPasswordShort},
		{"age too low", "a@b.com", "password123", "Full", 17, ErrAgeInvalid},
		{"age too high", "a@b.com", "password123", "Full", 91, ErrAgeInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Signup(tt.email, tt.password, tt.fullname, tt.age)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Signup() err = %v, want %v", err, tt.wantErr)
			}
		})
	}
	mock.ExpectationsWereMet()
}

func TestSignup_EmailExists(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)")).
		WithArgs("exists@test.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	_, err := svc.Signup("exists@test.com", "password123", "Full Name", 25)
	if !errors.Is(err, ErrEmailExists) {
		t.Errorf("Signup() err = %v, want ErrEmailExists", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestSignup_Success(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)")).
		WithArgs("new@test.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (email, password_hash, fullname, age, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id")).
		WithArgs("new@test.com", sqlmock.AnyArg(), "Full Name", 25).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO teams (user_id, name, country, budget, total_value, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id")).
		WithArgs(int64(1), "", "", int64(InitialBudget), int64(TotalPlayers*InitialPlayerValue)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))
	for i := 0; i < TotalPlayers; i++ {
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO players (team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at)")).
			WithArgs(int64(10), "Player", sqlmock.AnyArg(), "", sqlmock.AnyArg(), sqlmock.AnyArg(), InitialPlayerValue).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()
	// defer tx.Rollback() runs on return; in success path Commit already ran, Rollback is no-op

	resp, err := svc.Signup("new@test.com", "password123", "Full Name", 25)
	if err != nil {
		t.Fatalf("Signup() err = %v", err)
	}
	if resp == nil || resp.Email != "new@test.com" || resp.UserID != 1 || resp.Token == "" {
		t.Errorf("Signup() response = %+v", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- Login ---

func TestLogin_UserNotFound(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, fullname, age, password_hash, created_at, updated_at FROM users WHERE email = $1")).
		WithArgs("nobody@test.com").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Login("nobody@test.com", "any")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login() err = %v, want ErrInvalidCredentials", err)
	}
	mock.ExpectationsWereMet()
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	hash, _ := hashPasswordForTest("correct")
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, fullname, age, password_hash, created_at, updated_at FROM users WHERE email = $1")).
		WithArgs("user@test.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "fullname", "age", "password_hash", "created_at", "updated_at"}).
			AddRow(int64(1), "user@test.com", "User", 30, hash, now, now))

	_, err := svc.Login("user@test.com", "wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login() err = %v, want ErrInvalidCredentials", err)
	}
	mock.ExpectationsWereMet()
}

func TestLogin_Success(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	hash, _ := hashPasswordForTest("secret123")
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, fullname, age, password_hash, created_at, updated_at FROM users WHERE email = $1")).
		WithArgs("user@test.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "fullname", "age", "password_hash", "created_at", "updated_at"}).
			AddRow(int64(1), "user@test.com", "User Name", 28, hash, now, now))

	resp, err := svc.Login("user@test.com", "secret123")
	if err != nil {
		t.Fatalf("Login() err = %v", err)
	}
	if resp == nil || resp.Email != "user@test.com" || resp.UserID != 1 || resp.Token == "" {
		t.Errorf("Login() response = %+v", resp)
	}
	mock.ExpectationsWereMet()
}

// --- CreateUser ---

func TestCreateUser_Validation(t *testing.T) {
	svc, _ := newMockService(t)
	defer svc.Repo.Close()

	tests := []struct {
		name     string
		email    string
		password string
		fullname string
		age      int
		wantErr  error
	}{
		{"invalid email", "bad", "password123", "Full", 25, ErrEmailInvalid},
		{"short password", "a@b.com", "12345", "Full", 25, ErrPasswordShort},
		{"age negative", "a@b.com", "password123", "Full", -1, ErrAgeInvalid},
		{"age over 150", "a@b.com", "password123", "Full", 151, ErrAgeInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateUser(tt.email, tt.password, tt.fullname, tt.age)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CreateUser() err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateUser_EmailExists(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)")).
		WithArgs("taken@test.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	_, err := svc.CreateUser("taken@test.com", "password123", "Name", 25)
	if !errors.Is(err, ErrEmailExists) {
		t.Errorf("CreateUser() err = %v, want ErrEmailExists", err)
	}
	mock.ExpectationsWereMet()
}

func TestCreateUser_Success(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)")).
		WithArgs("newuser@test.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (email, password_hash, fullname, age, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id")).
		WithArgs("newuser@test.com", sqlmock.AnyArg(), "New User", 30).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, fullname, age, password_hash, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "fullname", "age", "password_hash", "created_at", "updated_at"}).
			AddRow(int64(42), "newuser@test.com", "New User", 30, "hash", now, now))

	u, err := svc.CreateUser("newuser@test.com", "password123", "New User", 30)
	if err != nil {
		t.Fatalf("CreateUser() err = %v", err)
	}
	if u == nil || u.ID != 42 || u.Email != "newuser@test.com" {
		t.Errorf("CreateUser() user = %+v", u)
	}
	mock.ExpectationsWereMet()
}

// --- GetTeam ---

func TestGetTeam_NotFound(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetTeam(999)
	if !errors.Is(err, ErrTeamNotFound) {
		t.Errorf("GetTeam() err = %v, want ErrTeamNotFound", err)
	}
	mock.ExpectationsWereMet()
}

func TestGetTeam_Success(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "My Team", "UK", int64(5000000), int64(0), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE team_id = $1 ORDER BY id")).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id", "first_name", "last_name", "country", "age", "position", "market_value", "created_at", "updated_at"}))

	team, err := svc.GetTeam(1)
	if err != nil {
		t.Fatalf("GetTeam() err = %v", err)
	}
	if team == nil || team.ID != 10 || team.UserID != 1 || team.TotalValue != 0 {
		t.Errorf("GetTeam() team = %+v", team)
	}
	mock.ExpectationsWereMet()
}

// --- CreateTeam ---

func TestCreateTeam_UserNotFound(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, fullname, age, password_hash, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.CreateTeam(999, "Team", "UK", 5000000)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("CreateTeam() err = %v, want ErrUserNotFound", err)
	}
	mock.ExpectationsWereMet()
}

func TestCreateTeam_AlreadyExists(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, fullname, age, password_hash, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "fullname", "age", "password_hash", "created_at", "updated_at"}).
			AddRow(int64(1), "u@t.com", "U", 25, "h", now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(5), int64(1), "Existing", "UK", int64(5000000), int64(20000000), now, now))

	_, err := svc.CreateTeam(1, "Team", "UK", 5000000)
	if !errors.Is(err, ErrTeamAlreadyExists) {
		t.Errorf("CreateTeam() err = %v, want ErrTeamAlreadyExists", err)
	}
	mock.ExpectationsWereMet()
}

// --- UpdateTeam ---

func TestUpdateTeam_NotFound(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := svc.UpdateTeam(999, "New", "US")
	if !errors.Is(err, ErrTeamNotFound) {
		t.Errorf("UpdateTeam() err = %v, want ErrTeamNotFound", err)
	}
	mock.ExpectationsWereMet()
}

func TestUpdateTeam_Success(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Old", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE teams SET name = $1, country = $2, updated_at = NOW() WHERE id = $3")).
		WithArgs("New Name", "US", int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.UpdateTeam(1, "New Name", "US")
	if err != nil {
		t.Fatalf("UpdateTeam() err = %v", err)
	}
	mock.ExpectationsWereMet()
}

// --- GetTeamPlayers ---

func TestGetTeamPlayers_NotFound(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetTeamPlayers(999)
	if !errors.Is(err, ErrTeamNotFound) {
		t.Errorf("GetTeamPlayers() err = %v, want ErrTeamNotFound", err)
	}
	mock.ExpectationsWereMet()
}

// --- CreatePlayer ---

func TestCreatePlayer_TeamNotFound(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.CreatePlayer(999, "First", "Last", "UK", "goalkeeper")
	if !errors.Is(err, ErrTeamNotFound) {
		t.Errorf("CreatePlayer() err = %v, want ErrTeamNotFound", err)
	}
	mock.ExpectationsWereMet()
}

func TestCreatePlayer_InvalidPosition(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Team", "UK", int64(5000000), int64(20000000), now, now))

	_, err := svc.CreatePlayer(1, "First", "Last", "UK", "striker")
	if !errors.Is(err, ErrInvalidPosition) {
		t.Errorf("CreatePlayer() err = %v, want ErrInvalidPosition", err)
	}
	mock.ExpectationsWereMet()
}

func TestCreatePlayer_Success(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Team", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO players (team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) RETURNING id")).
		WithArgs(int64(10), "First", "Last", "UK", sqlmock.AnyArg(), "midfielder", InitialPlayerValue).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id", "first_name", "last_name", "country", "age", "position", "market_value", "created_at", "updated_at"}).
			AddRow(int64(100), int64(10), "First", "Last", "UK", 22, "midfielder", InitialPlayerValue, now, now))

	p, err := svc.CreatePlayer(1, "First", "Last", "UK", "midfielder")
	if err != nil {
		t.Fatalf("CreatePlayer() err = %v", err)
	}
	if p == nil || p.ID != 100 || p.Position != "midfielder" {
		t.Errorf("CreatePlayer() player = %+v", p)
	}
	mock.ExpectationsWereMet()
}

// --- UpdatePlayer ---

func TestUpdatePlayer_TeamNotFound(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := svc.UpdatePlayer(999, 1, "A", "B", "UK")
	if !errors.Is(err, ErrTeamNotFound) {
		t.Errorf("UpdatePlayer() err = %v, want ErrTeamNotFound", err)
	}
	mock.ExpectationsWereMet()
}

func TestUpdatePlayer_PlayerNotFound(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Team", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := svc.UpdatePlayer(1, 999, "A", "B", "UK")
	if !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("UpdatePlayer() err = %v, want ErrPlayerNotFound", err)
	}
	mock.ExpectationsWereMet()
}

func TestUpdatePlayer_NotYourPlayer(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Team", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE id = $1")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id", "first_name", "last_name", "country", "age", "position", "market_value", "created_at", "updated_at"}).
			AddRow(int64(50), int64(99), "Other", "Player", "UK", 25, "defender", InitialPlayerValue, now, now))

	err := svc.UpdatePlayer(1, 50, "A", "B", "UK")
	if !errors.Is(err, ErrNotYourPlayer) {
		t.Errorf("UpdatePlayer() err = %v, want ErrNotYourPlayer", err)
	}
	mock.ExpectationsWereMet()
}

func TestUpdatePlayer_Success(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Team", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE id = $1")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id", "first_name", "last_name", "country", "age", "position", "market_value", "created_at", "updated_at"}).
			AddRow(int64(50), int64(10), "Old", "Name", "UK", 25, "defender", InitialPlayerValue, now, now))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE players SET first_name = $1, last_name = $2, country = $3, updated_at = NOW() WHERE id = $4")).
		WithArgs("New", "Name", "US", int64(50)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.UpdatePlayer(1, 50, "New", "Name", "US")
	if err != nil {
		t.Fatalf("UpdatePlayer() err = %v", err)
	}
	mock.ExpectationsWereMet()
}

// --- ListPlayerOnTransfer ---

func TestListPlayerOnTransfer_TeamNotFound(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := svc.ListPlayerOnTransfer(999, 1, 1000000)
	if !errors.Is(err, ErrTeamNotFound) {
		t.Errorf("ListPlayerOnTransfer() err = %v, want ErrTeamNotFound", err)
	}
	mock.ExpectationsWereMet()
}

func TestListPlayerOnTransfer_PlayerNotFound(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Team", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := svc.ListPlayerOnTransfer(1, 999, 1000000)
	if !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("ListPlayerOnTransfer() err = %v, want ErrPlayerNotFound", err)
	}
	mock.ExpectationsWereMet()
}

func TestListPlayerOnTransfer_NotYourPlayer(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Team", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE id = $1")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id", "first_name", "last_name", "country", "age", "position", "market_value", "created_at", "updated_at"}).
			AddRow(int64(50), int64(99), "Other", "Player", "UK", 25, "defender", InitialPlayerValue, now, now))

	err := svc.ListPlayerOnTransfer(1, 50, 1000000)
	if !errors.Is(err, ErrNotYourPlayer) {
		t.Errorf("ListPlayerOnTransfer() err = %v, want ErrNotYourPlayer", err)
	}
	mock.ExpectationsWereMet()
}

func TestListPlayerOnTransfer_InvalidPrice(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Team", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE id = $1")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id", "first_name", "last_name", "country", "age", "position", "market_value", "created_at", "updated_at"}).
			AddRow(int64(50), int64(10), "My", "Player", "UK", 25, "defender", InitialPlayerValue, now, now))

	err := svc.ListPlayerOnTransfer(1, 50, 0)
	if err == nil {
		t.Error("ListPlayerOnTransfer() expected error for price <= 0")
	}
	mock.ExpectationsWereMet()
}

func TestListPlayerOnTransfer_AlreadyListed(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Team", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE id = $1")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id", "first_name", "last_name", "country", "age", "position", "market_value", "created_at", "updated_at"}).
			AddRow(int64(50), int64(10), "My", "Player", "UK", 25, "defender", InitialPlayerValue, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, player_id, asking_price, listed_at FROM transfer_listings WHERE player_id = $1")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "player_id", "asking_price", "listed_at"}).AddRow(int64(1), int64(50), int64(2000000), now))

	err := svc.ListPlayerOnTransfer(1, 50, 1000000)
	if !errors.Is(err, ErrAlreadyListed) {
		t.Errorf("ListPlayerOnTransfer() err = %v, want ErrAlreadyListed", err)
	}
	mock.ExpectationsWereMet()
}

func TestListPlayerOnTransfer_Success(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Team", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE id = $1")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id", "first_name", "last_name", "country", "age", "position", "market_value", "created_at", "updated_at"}).
			AddRow(int64(50), int64(10), "My", "Player", "UK", 25, "defender", InitialPlayerValue, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, player_id, asking_price, listed_at FROM transfer_listings WHERE player_id = $1")).
		WithArgs(int64(50)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO transfer_listings (player_id, asking_price, listed_at) VALUES ($1, $2, NOW())")).
		WithArgs(int64(50), int64(1500000)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.ListPlayerOnTransfer(1, 50, 1500000)
	if err != nil {
		t.Fatalf("ListPlayerOnTransfer() err = %v", err)
	}
	mock.ExpectationsWereMet()
}

// --- GetTransferList ---

func TestGetTransferList_RepoError(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.id, t.player_id, t.asking_price, t.listed_at,")).
		WillReturnError(errors.New("db error"))

	_, err := svc.GetTransferList()
	if err == nil {
		t.Error("GetTransferList() expected error")
	}
	mock.ExpectationsWereMet()
}

func TestGetTransferList_Success(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT t.id, t.player_id, t.asking_price, t.listed_at").
		WillReturnRows(sqlmock.NewRows([]string{"t.id", "t.player_id", "t.asking_price", "t.listed_at", "p.id", "p.team_id", "p.first_name", "p.last_name", "p.country", "p.age", "p.position", "p.market_value", "p.created_at", "p.updated_at", "seller_team_name"}).
			AddRow(int64(1), int64(50), int64(2000000), now, int64(50), int64(10), "Listed", "Player", "UK", 25, "midfielder", InitialPlayerValue, now, now, ""))

	items, err := svc.GetTransferList()
	if err != nil {
		t.Fatalf("GetTransferList() err = %v", err)
	}
	if len(items) != 1 || items[0].PlayerID != 50 || items[0].AskingPrice != 2000000 {
		t.Errorf("GetTransferList() items = %+v", items)
	}
	mock.ExpectationsWereMet()
}

// --- BuyPlayer ---

func TestBuyPlayer_ListingNotFound(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, player_id, asking_price, listed_at FROM transfer_listings WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := svc.BuyPlayer(2, 999)
	if !errors.Is(err, ErrListingNotFound) {
		t.Errorf("BuyPlayer() err = %v, want ErrListingNotFound", err)
	}
	mock.ExpectationsWereMet()
}

func TestBuyPlayer_OwnPlayer(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, player_id, asking_price, listed_at FROM transfer_listings WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "player_id", "asking_price", "listed_at"}).AddRow(int64(1), int64(50), int64(2000000), now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE id = $1")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id", "first_name", "last_name", "country", "age", "position", "market_value", "created_at", "updated_at"}).
			AddRow(int64(50), int64(10), "P", "One", "UK", 25, "midfielder", InitialPlayerValue, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE id = $1")).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Seller", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(1), "Same", "UK", int64(5000000), int64(20000000), now, now))

	err := svc.BuyPlayer(1, 1)
	if !errors.Is(err, ErrCannotBuyOwnPlayer) {
		t.Errorf("BuyPlayer() err = %v, want ErrCannotBuyOwnPlayer", err)
	}
	mock.ExpectationsWereMet()
}

func TestBuyPlayer_InsufficientBudget(t *testing.T) {
	svc, mock := newMockService(t)
	defer svc.Repo.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, player_id, asking_price, listed_at FROM transfer_listings WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "player_id", "asking_price", "listed_at"}).AddRow(int64(1), int64(50), int64(10_000_000), now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, team_id, first_name, last_name, country, age, position, market_value, created_at, updated_at FROM players WHERE id = $1")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "team_id", "first_name", "last_name", "country", "age", "position", "market_value", "created_at", "updated_at"}).
			AddRow(int64(50), int64(10), "P", "One", "UK", 25, "midfielder", InitialPlayerValue, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE id = $1")).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(10), int64(99), "Seller", "UK", int64(5000000), int64(20000000), now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, name, country, budget, total_value, created_at, updated_at FROM teams WHERE user_id = $1")).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "country", "budget", "total_value", "created_at", "updated_at"}).
			AddRow(int64(20), int64(2), "Buyer", "UK", int64(1000000), int64(20000000), now, now))

	err := svc.BuyPlayer(2, 1)
	if !errors.Is(err, ErrInsufficientBudget) {
		t.Errorf("BuyPlayer() err = %v, want ErrInsufficientBudget", err)
	}
	mock.ExpectationsWereMet()
}

func hashPasswordForTest(password string) (string, error) {
	return auth.HashPassword(password)
}
