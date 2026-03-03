package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = errors.New("not found")

type DuelRepo struct {
	db *DB
}

func NewDuelRepo(db *DB) *DuelRepo {
	return &DuelRepo{db: db}
}

type User struct {
	ID       string
	Username string
	IsGuest  bool
}

type Duel struct {
	ID       string
	RoomCode string
	Theme    string
	LangFrom string
	LangTo   string
	Status   string
}

type Participant struct {
	ID     string
	DuelID string
	UserID string
}

type Round struct {
	ID        string
	DuelID    string
	Round     int
	PhraseID  string
	TimeLimit int
}

type DuelSummary struct {
	DuelID       string `json:"duel_id"`
	RoomCode     string `json:"room_code"`
	Status       string `json:"status"`
	StartedAt    string `json:"started_at"`
	FinishedAt   string `json:"finished_at"`
	WinnerUserID string `json:"winner_user_id"`
	CreatedAt    string `json:"created_at"`
	OpponentUserID   string `json:"opponent_user_id"`
	OpponentUsername string `json:"opponent_username"`
}

func (r *DuelRepo) CreateGuestUser(ctx context.Context, username string, ttlHours int) (*User, error) {
	if username == "" {
		return nil, errors.New("username required")
	}
	row := r.db.Pool.QueryRow(ctx,
		`INSERT INTO users (username, is_guest, guest_expires_at)
         VALUES ($1, true, now() + ($2::int * interval '1 hour'))
         RETURNING user_id`,
		username, ttlHours,
	)
	var id string
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return &User{ID: id, Username: username, IsGuest: true}, nil
}

func (r *DuelRepo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	row := r.db.Pool.QueryRow(ctx,
		`SELECT user_id, username, is_guest FROM users WHERE username = $1`,
		username,
	)
	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.IsGuest); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *DuelRepo) GetUserByID(ctx context.Context, userID string) (*User, error) {
	row := r.db.Pool.QueryRow(ctx,
		`SELECT user_id, username, is_guest FROM users WHERE user_id = $1`,
		userID,
	)
	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.IsGuest); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

type AuthUser struct {
	ID           string
	Username     string
	PasswordHash string
}

func (r *DuelRepo) CreateUser(ctx context.Context, username, email, passwordHash string) (*User, error) {
	row := r.db.Pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, is_guest)
         VALUES ($1, $2, $3, false)
         RETURNING user_id, username`,
		username, email, passwordHash,
	)
	var u User
	if err := row.Scan(&u.ID, &u.Username); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("user already exists")
		}
		return nil, err
	}
	return &u, nil
}

func (r *DuelRepo) GetAuthUserByUsernameOrEmail(ctx context.Context, login string) (*AuthUser, error) {
	row := r.db.Pool.QueryRow(ctx,
		`SELECT user_id, username, password_hash FROM users
         WHERE username = $1 OR email = $1`,
		login,
	)
	var u AuthUser
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *DuelRepo) CreateDuel(ctx context.Context, roomCode, createdByUserID, theme, langFrom, langTo string) (*Duel, error) {
	row := r.db.Pool.QueryRow(ctx,
		`INSERT INTO duels (room_code, created_by_user_id, theme, language_from, language_to)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING duel_id, room_code, theme, language_from, language_to, status`,
		roomCode, createdByUserID, theme, langFrom, langTo,
	)
	var d Duel
	if err := row.Scan(&d.ID, &d.RoomCode, &d.Theme, &d.LangFrom, &d.LangTo, &d.Status); err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DuelRepo) GetDuelByRoomCode(ctx context.Context, roomCode string) (*Duel, error) {
	row := r.db.Pool.QueryRow(ctx,
		`SELECT duel_id, room_code, theme, language_from, language_to, status
         FROM duels WHERE room_code = $1`,
		roomCode,
	)
	var d Duel
	if err := row.Scan(&d.ID, &d.RoomCode, &d.Theme, &d.LangFrom, &d.LangTo, &d.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *DuelRepo) EnsureParticipant(ctx context.Context, duelID, userID string, playerOrder int) (*Participant, error) {
	row := r.db.Pool.QueryRow(ctx,
		`INSERT INTO duel_participants (duel_id, user_id, player_order)
         VALUES ($1, $2, $3)
         ON CONFLICT (duel_id, user_id) DO UPDATE SET player_order = EXCLUDED.player_order
         RETURNING participant_id`,
		duelID, userID, playerOrder,
	)
	var id string
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return &Participant{ID: id, DuelID: duelID, UserID: userID}, nil
}

func (r *DuelRepo) CreateRound(ctx context.Context, duelID string, roundNumber int, phraseText, lang, topic string, timeLimitMs int) (*Round, error) {
	var phraseID string
	row := r.db.Pool.QueryRow(ctx,
		`SELECT phrase_id FROM phrases WHERE text = $1 AND lang = $2 AND topic = $3 LIMIT 1`,
		phraseText, lang, topic,
	)
	if err := row.Scan(&phraseID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			row = r.db.Pool.QueryRow(ctx,
				`INSERT INTO phrases (text, lang, topic) VALUES ($1, $2, $3) RETURNING phrase_id`,
				phraseText, lang, topic,
			)
			if err := row.Scan(&phraseID); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	row = r.db.Pool.QueryRow(ctx,
		`INSERT INTO game_rounds (duel_id, round_number, phrase_id, time_limit_ms)
         VALUES ($1, $2, $3, $4)
         ON CONFLICT (duel_id, round_number) DO UPDATE SET phrase_id = EXCLUDED.phrase_id
         RETURNING round_id`,
		duelID, roundNumber, phraseID, timeLimitMs,
	)
	var roundID string
	if err := row.Scan(&roundID); err != nil {
		return nil, err
	}
	return &Round{ID: roundID, DuelID: duelID, Round: roundNumber, PhraseID: phraseID, TimeLimit: timeLimitMs}, nil
}

func (r *DuelRepo) CreateAnswer(ctx context.Context, roundID, participantID, translationText string, correct bool, responseTimeMs, damageDealt int) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO player_answers (round_id, participant_id, translation_text, is_correct, response_time_ms, damage_dealt)
         VALUES ($1, $2, $3, $4, $5, $6)`,
		roundID, participantID, translationText, correct, responseTimeMs, damageDealt,
	)
	return err
}

func (r *DuelRepo) FinishDuel(ctx context.Context, duelID string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE duels SET status = 'finished', finished_at = now() WHERE duel_id = $1`,
		duelID,
	)
	return err
}

func (r *DuelRepo) MarkDuelStarted(ctx context.Context, duelID string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE duels SET status = 'active', started_at = now()
         WHERE duel_id = $1 AND started_at IS NULL`,
		duelID,
	)
	return err
}

func (r *DuelRepo) SetParticipantFinalHP(ctx context.Context, participantID string, finalHP int) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE duel_participants SET final_hp = $2 WHERE participant_id = $1`,
		participantID, finalHP,
	)
	return err
}

func (r *DuelRepo) SetDuelWinner(ctx context.Context, duelID, winnerUserID string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE duels SET winner_user_id = $2 WHERE duel_id = $1`,
		duelID, winnerUserID,
	)
	return err
}

func (r *DuelRepo) UpdateUserStats(ctx context.Context, userID string, won bool) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO user_statistics (user_id, total_duels_played, total_duels_won)
         VALUES ($1, 1, $2)
         ON CONFLICT (user_id)
         DO UPDATE SET
           total_duels_played = user_statistics.total_duels_played + 1,
           total_duels_won = user_statistics.total_duels_won + $2,
           updated_at = now()`,
		userID, boolToInt(won),
	)
	return err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

type UserStats struct {
	UserID            string  `json:"user_id"`
	TotalDuelsPlayed  int     `json:"total_duels_played"`
	TotalDuelsWon     int     `json:"total_duels_won"`
	TotalScore        int64   `json:"total_score"`
	OverallAccuracy   float64 `json:"overall_accuracy"`
	BestWinStreak     int     `json:"best_win_streak"`
	TotalPlayTimeMin  int     `json:"total_play_time_min"`
	UpdatedAt         string  `json:"updated_at"`
}

func (r *DuelRepo) GetUserStats(ctx context.Context, userID string) (*UserStats, error) {
	row := r.db.Pool.QueryRow(ctx,
		`SELECT user_id, total_duels_played, total_duels_won, total_score,
                overall_accuracy, best_win_streak, total_play_time_min, updated_at
         FROM user_statistics WHERE user_id = $1`,
		userID,
	)
	var s UserStats
	var updatedAt time.Time
	if err := row.Scan(&s.UserID, &s.TotalDuelsPlayed, &s.TotalDuelsWon, &s.TotalScore, &s.OverallAccuracy, &s.BestWinStreak, &s.TotalPlayTimeMin, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return empty stats for new users.
			return &UserStats{
				UserID:           userID,
				TotalDuelsPlayed: 0,
				TotalDuelsWon:    0,
				TotalScore:       0,
				OverallAccuracy:  0,
				BestWinStreak:    0,
				TotalPlayTimeMin: 0,
				UpdatedAt:        "",
			}, nil
		}
		return nil, err
	}
	s.UpdatedAt = updatedAt.Format(time.RFC3339)
	return &s, nil
}

func (r *DuelRepo) GetRecentDuels(ctx context.Context, userID string, limit int) ([]DuelSummary, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT d.duel_id, d.room_code, d.status, d.started_at, d.finished_at, d.winner_user_id, d.created_at,
		        u2.user_id, u2.username
         FROM duels d
         JOIN duel_participants p ON p.duel_id = d.duel_id
         LEFT JOIN duel_participants p2 ON p2.duel_id = d.duel_id AND p2.user_id <> p.user_id
         LEFT JOIN users u2 ON u2.user_id = p2.user_id
         WHERE p.user_id = $1
         ORDER BY d.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]DuelSummary, 0)
	for rows.Next() {
		var s DuelSummary
		var startedAt, finishedAt, createdAt *time.Time
		var winnerID *string
		var opponentID *string
		var opponentName *string
		if err := rows.Scan(&s.DuelID, &s.RoomCode, &s.Status, &startedAt, &finishedAt, &winnerID, &createdAt, &opponentID, &opponentName); err != nil {
			return nil, err
		}
		if winnerID != nil {
			s.WinnerUserID = *winnerID
		}
		if opponentID != nil {
			s.OpponentUserID = *opponentID
		}
		if opponentName != nil {
			s.OpponentUsername = *opponentName
		}
		if startedAt != nil {
			s.StartedAt = startedAt.Format(time.RFC3339)
		}
		if finishedAt != nil {
			s.FinishedAt = finishedAt.Format(time.RFC3339)
		}
		if createdAt != nil {
			s.CreatedAt = createdAt.Format(time.RFC3339)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
