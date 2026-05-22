package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"math/rand"
	"sort"
	"strings"
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
	ID              string
	Username        string
	IsGuest         bool
	Avatar          string
	Coins           int
	WinStreak       int
	UnlockedAvatars []string
}

type Duel struct {
	ID         string
	RoomCode   string
	Theme      string
	Difficulty int
	LangFrom   string
	LangTo     string
	Status     string
}

type Participant struct {
	ID          string
	DuelID      string
	UserID      string
	PlayerOrder int
}

type Round struct {
	ID        string
	DuelID    string
	Round     int
	PhraseID  string
	TimeLimit int
}

type DuelSummary struct {
	DuelID           string `json:"duel_id"`
	RoomCode         string `json:"room_code"`
	Status           string `json:"status"`
	StartedAt        string `json:"started_at"`
	FinishedAt       string `json:"finished_at"`
	WinnerUserID     string `json:"winner_user_id"`
	CreatedAt        string `json:"created_at"`
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
		`SELECT user_id, username, is_guest, COALESCE(avatar, 'default'), 
		 COALESCE(coins, 0), COALESCE(win_streak, 0), COALESCE(unlocked_avatars, '["default"]')
		 FROM users WHERE username = $1`,
		username,
	)
	var u User
	var unlockedStr string
	if err := row.Scan(&u.ID, &u.Username, &u.IsGuest, &u.Avatar, &u.Coins, &u.WinStreak, &unlockedStr); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal([]byte(unlockedStr), &u.UnlockedAvatars); err != nil {
		u.UnlockedAvatars = []string{"default"}
	}
	return &u, nil
}

func (r *DuelRepo) GetUserByID(ctx context.Context, userID string) (*User, error) {
	row := r.db.Pool.QueryRow(ctx,
		`SELECT user_id, username, is_guest, COALESCE(avatar, 'default'), 
		 COALESCE(coins, 0), COALESCE(win_streak, 0), COALESCE(unlocked_avatars, '["default"]')
		 FROM users WHERE user_id = $1`,
		userID,
	)
	var u User
	var unlockedStr string
	if err := row.Scan(&u.ID, &u.Username, &u.IsGuest, &u.Avatar, &u.Coins, &u.WinStreak, &unlockedStr); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal([]byte(unlockedStr), &u.UnlockedAvatars); err != nil {
		u.UnlockedAvatars = []string{"default"}
	}
	return &u, nil
}

func (r *DuelRepo) BuyAvatar(ctx context.Context, userID, avatarID string, price int) error {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Coins < price {
		return errors.New("not enough coins")
	}

	for _, a := range user.UnlockedAvatars {
		if a == avatarID {
			return errors.New("avatar already unlocked")
		}
	}

	unlocked := append(user.UnlockedAvatars, avatarID)
	unlockedJSON, _ := json.Marshal(unlocked)

	_, err = r.db.Pool.Exec(ctx,
		`UPDATE users SET coins = coins - $1, unlocked_avatars = $2 WHERE user_id = $3`,
		price, string(unlockedJSON), userID,
	)
	return err
}

func (r *DuelRepo) AddCoins(ctx context.Context, userID string, amount int) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE users SET coins = coins + $1 WHERE user_id = $2`,
		amount, userID,
	)
	return err
}

func (r *DuelRepo) UpdateWinStreak(ctx context.Context, userID string, won bool) error {
	if won {
		_, err := r.db.Pool.Exec(ctx,
			`UPDATE users SET win_streak = win_streak + 1 WHERE user_id = $1`,
			userID,
		)
		return err
	}
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE users SET win_streak = 0 WHERE user_id = $1`,
		userID,
	)
	return err
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

func (r *DuelRepo) ConvertGuestToUser(ctx context.Context, username, email, passwordHash string) (*User, error) {
	row := r.db.Pool.QueryRow(ctx,
		`UPDATE users SET email = $2, password_hash = $3, is_guest = false, guest_expires_at = NULL
         WHERE username = $1 AND is_guest = true
         RETURNING user_id, username`,
		username, email, passwordHash,
	)
	var u User
	if err := row.Scan(&u.ID, &u.Username); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
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

func (r *DuelRepo) UpdateUsername(ctx context.Context, userID, newUsername string) error {
	result, err := r.db.Pool.Exec(ctx,
		`UPDATE users SET username = $2 WHERE user_id = $1`,
		userID, newUsername,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.New("username already taken")
		}
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *DuelRepo) UpdateAvatar(ctx context.Context, userID, newAvatar string) error {
	result, err := r.db.Pool.Exec(ctx,
		`UPDATE users SET avatar = $2 WHERE user_id = $1`,
		userID, newAvatar,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *DuelRepo) CreateDuel(ctx context.Context, roomCode, createdByUserID, theme string, difficulty int, langFrom, langTo string) (*Duel, error) {
	row := r.db.Pool.QueryRow(ctx,
		`INSERT INTO duels (room_code, created_by_user_id, theme, difficulty, language_from, language_to)
         VALUES ($1, $2, $3, $4, $5, $6)
         RETURNING duel_id, room_code, theme, difficulty, language_from, language_to, status`,
		roomCode, createdByUserID, theme, difficulty, langFrom, langTo,
	)
	var d Duel
	if err := row.Scan(&d.ID, &d.RoomCode, &d.Theme, &d.Difficulty, &d.LangFrom, &d.LangTo, &d.Status); err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DuelRepo) GetDuelByRoomCode(ctx context.Context, roomCode string) (*Duel, error) {
	row := r.db.Pool.QueryRow(ctx,
		`SELECT duel_id, room_code, theme, difficulty, language_from, language_to, status
         FROM duels WHERE room_code = $1 AND status != 'finished'
         ORDER BY created_at DESC LIMIT 1`,
		roomCode,
	)
	var d Duel
	if err := row.Scan(&d.ID, &d.RoomCode, &d.Theme, &d.Difficulty, &d.LangFrom, &d.LangTo, &d.Status); err != nil {
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
         RETURNING participant_id, duel_id, user_id, player_order`,
		duelID, userID, playerOrder,
	)
	var p Participant
	if err := row.Scan(&p.ID, &p.DuelID, &p.UserID, &p.PlayerOrder); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *DuelRepo) DeletePendingDuel(ctx context.Context, roomCode string) error {
	_, err := r.db.Pool.Exec(ctx,
		`DELETE FROM duels WHERE room_code = $1 AND status = 'pending'`,
		roomCode,
	)
	return err
}

func (r *DuelRepo) CreateRound(ctx context.Context, duelID string, roundNumber int, phraseText, correctAnswer, lang, topic string, timeLimitMs int, validAnswers []string) (*Round, error) {
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

	if validAnswers == nil {
		validAnswers = []string{}
	}

	row = r.db.Pool.QueryRow(ctx,
		`INSERT INTO game_rounds (duel_id, round_number, phrase_id, correct_answer, valid_answers, time_limit_ms)
         VALUES ($1, $2, $3, $4, $5, $6)
         ON CONFLICT (duel_id, round_number) DO UPDATE SET phrase_id = EXCLUDED.phrase_id, correct_answer = EXCLUDED.correct_answer, valid_answers = EXCLUDED.valid_answers
         RETURNING round_id`,
		duelID, roundNumber, phraseID, correctAnswer, validAnswers, timeLimitMs,
	)
	var roundID string
	if err := row.Scan(&roundID); err != nil {
		return nil, err
	}
	return &Round{ID: roundID, DuelID: duelID, Round: roundNumber, PhraseID: phraseID, TimeLimit: timeLimitMs}, nil
}

// GetRoundID returns the DB round_id for a given duel and round number.
func (r *DuelRepo) GetRoundID(ctx context.Context, duelID string, roundNumber int) (string, error) {
	var id string
	err := r.db.Pool.QueryRow(ctx,
		`SELECT round_id::text FROM game_rounds WHERE duel_id = $1 AND round_number = $2 LIMIT 1`,
		duelID, roundNumber,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetParticipantID returns the DB participant_id for a given duel and user.
func (r *DuelRepo) GetParticipantID(ctx context.Context, duelID, userID string) (string, error) {
	var id string
	err := r.db.Pool.QueryRow(ctx,
		`SELECT participant_id::text FROM duel_participants WHERE duel_id = $1 AND user_id = $2 LIMIT 1`,
		duelID, userID,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *DuelRepo) CreateAnswer(ctx context.Context, roundID, participantID, translationText string, correct bool, responseTimeMs, damageDealt int) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO player_answers (round_id, participant_id, translation_text, is_correct, response_time_ms, damage_dealt)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (round_id, participant_id) DO NOTHING`,
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
	UserID           string  `json:"user_id"`
	TotalDuelsPlayed int     `json:"total_duels_played"`
	TotalDuelsWon    int     `json:"total_duels_won"`
	TotalScore       int64   `json:"total_score"`
	OverallAccuracy  float64 `json:"overall_accuracy"`
	BestWinStreak    int     `json:"best_win_streak"`
	TotalPlayTimeMin int     `json:"total_play_time_min"`
	UpdatedAt        string  `json:"updated_at"`
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

// Rating constants
const (
	RatingWin             = 25
	RatingLoss            = -15
	RatingMin             = 0
	RatingDefault         = 1000
	MaxStreakLossesAtZero = 10
)

// Rating represents user rating data
type Rating struct {
	UserID            string `json:"user_id"`
	Username          string `json:"username"`
	Avatar            string `json:"avatar"`
	Elo               int    `json:"elo"`
	Rank              string `json:"rank"`
	GamesPlayed       int    `json:"games_played"`
	Wins              int    `json:"wins"`
	Losses            int    `json:"losses"`
	CurrentStreak     int    `json:"current_streak"`
	BestStreak        int    `json:"best_streak"`
	Coins             int    `json:"coins"`
	XP                int64  `json:"xp"`
	Level             int    `json:"level"`
	TotalLossesAtZero int    `json:"-"`
}

type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Elo      int    `json:"elo"`
	RankTier string `json:"rank_tier"`
	RankName string `json:"rank_name"`
	Games    int    `json:"games_played"`
}

// UpdateRating updates ratings for winner and loser
func (r *DuelRepo) UpdateRating(ctx context.Context, winnerID, loserID string) error {
	// Get current ratings
	winnerRating, _ := r.GetUserRating(ctx, winnerID)
	loserRating, _ := r.GetUserRating(ctx, loserID)

	// Default values if ratings don't exist
	currentWinnerStreak := 0
	loserLossesAtZero := 0

	if winnerRating != nil {
		currentWinnerStreak = winnerRating.CurrentStreak
	}
	if loserRating != nil {
		loserLossesAtZero = loserRating.TotalLossesAtZero
	}

	newWinnerElo := RatingDefault + RatingWin
	newLoserElo := RatingDefault + RatingLoss
	if loserRating != nil {
		newWinnerElo = winnerRating.Elo + RatingWin
		newLoserElo = loserRating.Elo + RatingLoss
	}
	if newLoserElo < RatingMin {
		newLoserElo = RatingMin
	}

	// Update winner
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO user_ratings (user_id, elo, games_played, wins, current_streak, best_streak)
		VALUES ($1, $2, 1, 1, 1, GREATEST($3, 1))
		ON CONFLICT (user_id) DO UPDATE SET
			elo = $2,
			games_played = user_ratings.games_played + 1,
			wins = user_ratings.wins + 1,
			current_streak = user_ratings.current_streak + 1,
			best_streak = GREATEST(user_ratings.best_streak, user_ratings.current_streak + 1),
			updated_at = NOW()`,
		winnerID, newWinnerElo, currentWinnerStreak+1)
	if err != nil {
		return err
	}

	// Update loser
	lossesAtZero := 0
	if newLoserElo == 0 {
		lossesAtZero = loserLossesAtZero + 1
	}

	_, err = r.db.Pool.Exec(ctx, `
		INSERT INTO user_ratings (user_id, elo, games_played, losses, current_streak, total_losses_at_zero)
		VALUES ($1, $2, 1, 1, -1, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			elo = $2,
			games_played = user_ratings.games_played + 1,
			losses = user_ratings.losses + 1,
			current_streak = CASE WHEN user_ratings.elo = 0 THEN user_ratings.current_streak - 1 ELSE -1 END,
			total_losses_at_zero = $3,
			updated_at = NOW()`,
		loserID, newLoserElo, lossesAtZero)
	if err != nil {
		return err
	}

	// Update ranks
	if err := r.updateRankForUser(ctx, winnerID); err != nil {
		return err
	}
	if err := r.updateRankForUser(ctx, loserID); err != nil {
		return err
	}

	// Award coins for winner
	winCoins := 10
	_, err = r.db.Pool.Exec(ctx, `
		UPDATE users SET coins = COALESCE(coins, 0) + $1 WHERE user_id = $2`,
		winCoins, winnerID)
	if err != nil {
		log.Printf("Award coins error: %v", err)
	}

	return nil
}

func (r *DuelRepo) updateRankForUser(ctx context.Context, userID string) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE user_ratings SET rank = (
			CASE
				WHEN elo >= 3000 THEN 'master'
				WHEN elo >= 2000 THEN 'expert'
				WHEN elo >= 1000 THEN 'apprentice'
				WHEN elo > 0 OR total_losses_at_zero <= $2 THEN 'newbie'
				ELSE 'struggler'
			END
		), updated_at = NOW()
		WHERE user_id = $1`,
		userID, MaxStreakLossesAtZero)
	return err
}

// GetUserRating returns rating for a user
func (r *DuelRepo) GetUserRating(ctx context.Context, userID string) (*Rating, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT r.user_id, u.username, COALESCE(u.avatar, 'default'),
			   r.elo, r.rank, r.games_played, r.wins, r.losses,
			   r.current_streak, r.best_streak, COALESCE(r.total_losses_at_zero, 0),
			   COALESCE(r.coins, 0), COALESCE(r.xp, 0)
		FROM user_ratings r
		JOIN users u ON u.user_id = r.user_id
		WHERE r.user_id = $1`, userID)

	var rating Rating
	var totalLossesAtZero int
	err := row.Scan(&rating.UserID, &rating.Username, &rating.Avatar,
		&rating.Elo, &rating.Rank, &rating.GamesPlayed, &rating.Wins, &rating.Losses,
		&rating.CurrentStreak, &rating.BestStreak, &totalLossesAtZero, &rating.Coins, &rating.XP)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Rating{
				UserID: userID, Elo: RatingDefault, Rank: "newbie",
				GamesPlayed: 0, Wins: 0, Losses: 0,
				CurrentStreak: 0, BestStreak: 0,
			}, nil
		}
		return nil, err
	}
	// Calculate level: floor(sqrt(xp / 100))
	rating.Level = calculateLevel(rating.XP)
	return &rating, nil
}

// calculateLevel returns level based on XP using formula: floor(sqrt(xp / 100))
func calculateLevel(xp int64) int {
	if xp <= 0 {
		return 1
	}
	level := int(math.Floor(math.Sqrt(float64(xp) / 100)))
	if level < 1 {
		return 1
	}
	return level + 1 // Start from level 1
}

// GetLeaderboard returns top players (excluding guests)
func (r *DuelRepo) GetLeaderboard(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT r.user_id, u.username, COALESCE(u.avatar, 'default'),
			   r.elo, r.rank, r.games_played,
			   CASE
				   WHEN r.elo >= 3000 THEN '💎 Master'
				   WHEN r.elo >= 2000 THEN '🥇 Expert'
				   WHEN r.elo >= 1000 THEN '🥈 Apprentice'
				   WHEN r.total_losses_at_zero > $2 THEN '😔 Struggler'
				   ELSE '🥉 Newbie'
			   END as rank_name
		FROM user_ratings r
		JOIN users u ON u.user_id = r.user_id
		WHERE u.is_guest = false AND r.games_played > 0
		ORDER BY r.elo DESC
		LIMIT $1`, limit, MaxStreakLossesAtZero)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	rank := 1
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.UserID, &entry.Username, &entry.Avatar,
			&entry.Elo, &entry.RankTier, &entry.Games, &entry.RankName); err != nil {
			return nil, err
		}
		entry.Rank = rank
		entries = append(entries, entry)
		rank++
	}
	return entries, rows.Err()
}

// EnsureRating creates rating entry if not exists
func (r *DuelRepo) EnsureRating(ctx context.Context, userID string) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO user_ratings (user_id, elo, rank)
		VALUES ($1, $2, 'newbie')
		ON CONFLICT (user_id) DO NOTHING`, userID, RatingDefault)
	return err
}

// Achievement represents a user achievement
type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	XPReward    int    `json:"xp_reward"`
	CoinsReward int    `json:"coins_reward"`
	Unlocked    bool   `json:"unlocked"`
	UnlockedAt  string `json:"unlocked_at,omitempty"`
}

// UserAchievement represents unlocked achievement
type UserAchievement struct {
	AchievementID string `json:"achievement_id"`
	UnlockedAt    string `json:"unlocked_at"`
}

// GetAllAchievements returns all achievements with unlock status for user
func (r *DuelRepo) GetAllAchievements(ctx context.Context, userID string) ([]Achievement, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT a.id, a.name, a.description, a.icon, a.xp_reward, COALESCE(a.coins_reward, 0),
			   ua.unlocked_at::text
		FROM achievements a
		LEFT JOIN user_achievements ua ON a.id = ua.achievement_id AND ua.user_id = $1
		ORDER BY a.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []Achievement
	for rows.Next() {
		var a Achievement
		var unlockedAt *string
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.Icon, &a.XPReward, &a.CoinsReward, &unlockedAt); err != nil {
			return nil, err
		}
		a.Unlocked = unlockedAt != nil
		if unlockedAt != nil {
			a.UnlockedAt = *unlockedAt
		}
		achievements = append(achievements, a)
	}
	return achievements, rows.Err()
}

// GetUnlockedAchievements returns only unlocked achievements for user
func (r *DuelRepo) GetUnlockedAchievements(ctx context.Context, userID string) ([]UserAchievement, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT achievement_id, unlocked_at
		FROM user_achievements
		WHERE user_id = $1
		ORDER BY unlocked_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []UserAchievement
	for rows.Next() {
		var a UserAchievement
		if err := rows.Scan(&a.AchievementID, &a.UnlockedAt); err != nil {
			return nil, err
		}
		achievements = append(achievements, a)
	}
	return achievements, rows.Err()
}

// UnlockAchievement unlocks an achievement for user
func (r *DuelRepo) UnlockAchievement(ctx context.Context, userID, achievementID string) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO user_achievements (user_id, achievement_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, achievement_id) DO NOTHING`, userID, achievementID)
	return err
}

// AwardXP adds XP to user's rating and returns old and new level
func (r *DuelRepo) AwardXP(ctx context.Context, userID string, amount int) (oldLevel, newLevel int, err error) {
	// Ensure column exists
	_, _ = r.db.Pool.Exec(ctx, `ALTER TABLE user_ratings ADD COLUMN IF NOT EXISTS xp BIGINT DEFAULT 0`)

	// Ensure user_ratings row exists
	_, _ = r.db.Pool.Exec(ctx, `
		INSERT INTO user_ratings (user_id, elo, rank, xp)
		VALUES ($1, 1000, 'newbie', 0)
		ON CONFLICT (user_id) DO NOTHING`, userID)

	// Get current XP
	var currentXP int64
	r.db.Pool.QueryRow(ctx, `SELECT COALESCE(xp, 0) FROM user_ratings WHERE user_id = $1`, userID).Scan(&currentXP)
	oldLevel = calculateLevel(currentXP)

	// Add XP
	_, err = r.db.Pool.Exec(ctx, `
		UPDATE user_ratings SET xp = COALESCE(xp, 0) + $1 WHERE user_id = $2`, amount, userID)
	if err != nil {
		return oldLevel, oldLevel, err
	}

	// Get new XP and level
	var newXP int64
	r.db.Pool.QueryRow(ctx, `SELECT COALESCE(xp, 0) FROM user_ratings WHERE user_id = $1`, userID).Scan(&newXP)
	newLevel = calculateLevel(newXP)

	log.Printf("Awarded %d XP to user %s: level %d -> %d", amount, userID, oldLevel, newLevel)
	return oldLevel, newLevel, nil
}

// AwardCoins adds coins to user's rating
func (r *DuelRepo) AwardCoins(ctx context.Context, userID string, amount int) error {
	// First ensure the column exists
	_, _ = r.db.Pool.Exec(ctx, `ALTER TABLE user_ratings ADD COLUMN IF NOT EXISTS coins INT DEFAULT 0`)

	// Then award coins
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE user_ratings SET coins = COALESCE(coins, 0) + $1 WHERE user_id = $2`, amount, userID)
	return err
}

// AwardCoinsForUnlockedAchievements retroactively awards coins for already unlocked achievements
func (r *DuelRepo) AwardCoinsForUnlockedAchievements(ctx context.Context, userID string) (int, error) {
	log.Printf("AwardCoinsForUnlockedAchievements called for user: %s", userID)

	// Sync achievements based on user stats first
	synced, err := r.SyncAchievementsFromStats(ctx, userID)
	if err != nil {
		log.Printf("SyncAchievementsFromStats error: %v", err)
	}
	log.Printf("Synced %d achievements for user %s", synced, userID)

	// Count unlocked achievements
	var achCount int
	err = r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM user_achievements WHERE user_id = $1`, userID).Scan(&achCount)
	if err != nil {
		log.Printf("Count user_achievements error: %v", err)
		return 0, err
	}
	log.Printf("User %s has %d unlocked achievements", userID, achCount)

	if achCount == 0 {
		return 0, nil
	}

	// Get total coins from unlocked achievements
	var totalCoins int
	err = r.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(COALESCE(coins_reward, 0)), 0)
		FROM achievements
		WHERE id IN (SELECT achievement_id FROM user_achievements WHERE user_id = $1)`, userID).Scan(&totalCoins)
	if err != nil {
		log.Printf("Sum coins_reward error: %v, using fallback", err)
		// Fallback: default coin values
		totalCoins = achCount * 10
	}
	log.Printf("Calculated total coins: %d", totalCoins)

	// Update coins in users table (where BuyAvatar reads from)
	if totalCoins > 0 {
		_, err = r.db.Pool.Exec(ctx, `
			UPDATE users SET coins = COALESCE(coins, 0) + $1 WHERE user_id = $2`, totalCoins, userID)
		if err != nil {
			log.Printf("Update coins error: %v", err)
			return 0, err
		}
		log.Printf("Successfully awarded %d coins to user %s in users table", totalCoins, userID)
	}

	return totalCoins, nil
}

// SyncAchievementsFromStats syncs achievements based on user statistics
func (r *DuelRepo) SyncAchievementsFromStats(ctx context.Context, userID string) (int, error) {
	// First ensure achievements table has data
	if err := r.EnsureAchievementsExist(ctx); err != nil {
		log.Printf("EnsureAchievementsExist error: %v", err)
	}

	// Get user stats
	var wins, games int
	row := r.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(total_duels_won, 0), COALESCE(total_duels_played, 0)
		FROM user_statistics WHERE user_id = $1`, userID)
	if err := row.Scan(&wins, &games); err != nil {
		log.Printf("SyncAchievementsFromStats - no stats for user: %v", err)
		return 0, nil
	}

	// Get best streak
	var bestStreak int
	r.db.Pool.QueryRow(ctx, `SELECT COALESCE(best_win_streak, 0) FROM user_statistics WHERE user_id = $1`, userID).Scan(&bestStreak)

	log.Printf("SyncAchievementsFromStats - user stats: wins=%d, games=%d, bestStreak=%d", wins, games, bestStreak)

	// Define which achievements should be unlocked
	type achCheck struct {
		id        string
		unlockWin int // required wins, 0 = don't check wins
		unlockGam int // required games, 0 = don't check games
		unlockStr int // required streak, 0 = don't check streak
	}
	achChecks := []achCheck{
		{"first_win", 1, 0, 0},
		{"warrior", 10, 0, 0},
		{"veteran", 50, 0, 0},
		{"champion", 100, 0, 0},
		{"streak_5", 0, 0, 5},
		{"streak_10", 0, 0, 10},
		{"games_10", 0, 10, 0},
		{"games_50", 0, 50, 0},
	}

	synced := 0
	for _, c := range achChecks {
		shouldUnlock := false
		if c.unlockWin > 0 && wins >= c.unlockWin {
			shouldUnlock = true
			log.Printf("Achievement %s qualifies: wins=%d >= %d", c.id, wins, c.unlockWin)
		}
		if c.unlockGam > 0 && games >= c.unlockGam {
			shouldUnlock = true
			log.Printf("Achievement %s qualifies: games=%d >= %d", c.id, games, c.unlockGam)
		}
		if c.unlockStr > 0 && bestStreak >= c.unlockStr {
			shouldUnlock = true
			log.Printf("Achievement %s qualifies: bestStreak=%d >= %d", c.id, bestStreak, c.unlockStr)
		}

		if shouldUnlock {
			result, err := r.db.Pool.Exec(ctx, `
				INSERT INTO user_achievements (user_id, achievement_id)
				VALUES ($1, $2)
				ON CONFLICT (user_id, achievement_id) DO NOTHING`, userID, c.id)
			if err != nil {
				log.Printf("INSERT user_achievements error: %v", err)
			} else {
				rowsAffected := result.RowsAffected()
				log.Printf("Achievement %s insert result: rows_affected=%d", c.id, rowsAffected)
				if rowsAffected > 0 {
					synced++
				}
			}
		}
	}

	return synced, nil
}

// EnsureAchievementsExist creates base achievements if they don't exist
func (r *DuelRepo) EnsureAchievementsExist(ctx context.Context) error {
	// Create achievements table if not exists
	_, err := r.db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS achievements (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			icon VARCHAR(50),
			xp_reward INT DEFAULT 0,
			coins_reward INT DEFAULT 0
		)`)
	if err != nil {
		log.Printf("Create achievements table error: %v", err)
	}

	// Create user_achievements table if not exists
	_, err = r.db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user_achievements (
			user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
			achievement_id VARCHAR(50) REFERENCES achievements(id),
			unlocked_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (user_id, achievement_id)
		)`)
	if err != nil {
		log.Printf("Create user_achievements table error: %v", err)
	}

	// Insert base achievements (ignore duplicates)
	achievements := []struct {
		id          string
		name        string
		description string
		icon        string
		xpReward    int
		coinsReward int
	}{
		{"first_win", "Первая победа", "Выиграйте первый матч", "🏆", 10, 5},
		{"warrior", "Воин", "Выиграйте 10 матчей", "⚔️", 25, 15},
		{"veteran", "Ветеран", "Выиграйте 50 матчей", "🛡️", 50, 30},
		{"champion", "Чемпион", "Выиграйте 100 матчей", "👑", 100, 75},
		{"streak_5", "Натиск", "5 побед подряд", "🔥", 20, 10},
		{"streak_10", "Мастер натиска", "10 побед подряд", "💥", 50, 35},
		{"games_10", "Новичок", "Сыграйте 10 матчей", "🎮", 10, 5},
		{"games_50", "Игрок", "Сыграйте 50 матчей", "🎯", 25, 20},
	}

	for _, a := range achievements {
		_, err := r.db.Pool.Exec(ctx, `
			INSERT INTO achievements (id, name, description, icon, xp_reward, coins_reward)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				icon = EXCLUDED.icon,
				xp_reward = EXCLUDED.xp_reward,
				coins_reward = EXCLUDED.coins_reward`, a.id, a.name, a.description, a.icon, a.xpReward, a.coinsReward)
		if err != nil {
			log.Printf("Insert achievement %s error: %v", a.id, err)
		}
	}

	log.Printf("Ensured achievements exist")
	return nil
}

// CheckAndUnlockAchievements checks and unlocks achievements based on user stats
func (r *DuelRepo) CheckAndUnlockAchievements(ctx context.Context, userID string, isWinner bool, currentStreak int) ([]Achievement, error) {
	// Get user stats directly from user_statistics
	var wins, games int
	row := r.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(total_duels_won, 0), COALESCE(total_duels_played, 0)
		FROM user_statistics WHERE user_id = $1`, userID)
	if err := row.Scan(&wins, &games); err != nil {
		// If no stats yet, start from 0
		wins = 0
		games = 0
	}

	// Add current game's contribution if winner
	if isWinner {
		wins++
	}
	games++

	// Define achievement checks
	type check struct {
		id      string
		enabled bool
	}
	checks := []check{
		{"first_win", isWinner && wins >= 1},
		{"warrior", wins >= 10},
		{"veteran", wins >= 50},
		{"champion", wins >= 100},
		{"streak_5", isWinner && currentStreak >= 5},
		{"streak_10", isWinner && currentStreak >= 10},
		{"games_10", games >= 10},
		{"games_50", games >= 50},
	}

	var unlocked []Achievement
	for _, c := range checks {
		if c.enabled {
			err := r.UnlockAchievement(ctx, userID, c.id)
			if err != nil {
				log.Printf("UnlockAchievement error: %v", err)
				continue
			}
			// Get achievement details - try with coins_reward first, fallback without
			row := r.db.Pool.QueryRow(ctx, `
				SELECT id, name, description, icon, xp_reward, COALESCE(coins_reward, 0)
				FROM achievements WHERE id = $1`, c.id)
			var a Achievement
			if err := row.Scan(&a.ID, &a.Name, &a.Description, &a.Icon, &a.XPReward, &a.CoinsReward); err != nil {
				log.Printf("Get achievement %s error: %v", c.id, err)
				continue
			}
			a.Unlocked = true
			// Award coins for this achievement
			if a.CoinsReward > 0 {
				if err := r.AwardCoins(ctx, userID, a.CoinsReward); err != nil {
					log.Printf("AwardCoins error: %v", err)
				} else {
					log.Printf("Awarded %d coins to user %s for achievement %s", a.CoinsReward, userID, a.ID)
				}
			}
			unlocked = append(unlocked, a)
		}
	}

	return unlocked, nil
}

// AIPhrase represents an AI-generated phrase
type AIPhrase struct {
	ID         string
	DuelID     sql.NullString
	Prompt     string
	Answers    []string
	Topic      string
	Difficulty string
	LangFrom   string
	LangTo     string
	Used       bool
	CreatedAt  string
}

// SaveAIPhrase saves an AI-generated phrase
func (r *DuelRepo) SaveAIPhrase(ctx context.Context, duelID, roomCode, prompt string, answers []string, topic, difficulty, langFrom, langTo string) error {
	// Use NULL if duelID is empty
	if duelID == "" {
		_, err := r.db.Pool.Exec(ctx, `
			INSERT INTO ai_phrases (room_code, prompt, answers, topic, difficulty, lang_from, lang_to)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			roomCode, prompt, answers, topic, difficulty, langFrom, langTo)
		return err
	}
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO ai_phrases (duel_id, room_code, prompt, answers, topic, difficulty, lang_from, lang_to)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		duelID, roomCode, prompt, answers, topic, difficulty, langFrom, langTo)
	return err
}

// DeleteAIPhrasesByRoomCode deletes all AI phrases for a room (for regeneration)
func (r *DuelRepo) DeleteAIPhrasesByRoomCode(ctx context.Context, roomCode string) (int64, error) {
	tag, err := r.db.Pool.Exec(ctx, `DELETE FROM ai_phrases WHERE room_code = $1`, roomCode)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// GetAIPhrases returns AI-generated phrases for a duel or room.
// Fetches all unused phrases (no SQL LIMIT) so Go-level dedup sees the full set,
// then shuffles and caps at 20. Ordered by created_at DESC so newest batch wins.
func (r *DuelRepo) GetAIPhrases(ctx context.Context, duelID, roomCode string) ([]AIPhrase, error) {
	var rows pgx.Rows
	var err error

	if duelID == "" {
		rows, err = r.db.Pool.Query(ctx, `
			SELECT phrase_id, duel_id, prompt, answers, topic, difficulty, lang_from, lang_to, used, created_at
			FROM ai_phrases
			WHERE used = false AND room_code = $1
			ORDER BY created_at DESC`, roomCode)
	} else {
		rows, err = r.db.Pool.Query(ctx, `
			SELECT phrase_id, duel_id, prompt, answers, topic, difficulty, lang_from, lang_to, used, created_at
			FROM ai_phrases
			WHERE used = false AND (duel_id = $1 OR room_code = $2)
			ORDER BY created_at DESC`, duelID, roomCode)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Deduplicate by prompt (case-insensitive), preferring newest rows (already sorted DESC)
	seenPrompts := make(map[string]bool)
	var all []AIPhrase
	for rows.Next() {
		var p AIPhrase
		var createdAt time.Time
		if err := rows.Scan(&p.ID, &p.DuelID, &p.Prompt, &p.Answers, &p.Topic, &p.Difficulty,
			&p.LangFrom, &p.LangTo, &p.Used, &createdAt); err != nil {
			return nil, err
		}
		p.CreatedAt = createdAt.Format(time.RFC3339)
		promptLower := strings.ToLower(strings.TrimSpace(p.Prompt))
		if !seenPrompts[promptLower] {
			seenPrompts[promptLower] = true
			all = append(all, p)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Shuffle for game variety, then cap at 20
	rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	if len(all) > 20 {
		all = all[:20]
	}
	return all, nil
}

// MarkAIPhraseUsed marks an AI phrase as used
func (r *DuelRepo) MarkAIPhraseUsed(ctx context.Context, phraseID string) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE ai_phrases SET used = true WHERE phrase_id = $1`, phraseID)
	return err
}

// EnsureAIPhraseTable creates the ai_phrases table if not exists
func (r *DuelRepo) EnsureAIPhraseTable(ctx context.Context) error {
	_, err := r.db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS ai_phrases (
			phrase_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			duel_id UUID REFERENCES duels(duel_id) ON DELETE CASCADE,
			prompt VARCHAR(255) NOT NULL,
			answers TEXT[] NOT NULL,
			topic VARCHAR(30) NOT NULL,
			difficulty VARCHAR(20) NOT NULL DEFAULT 'intermediate',
			lang_from VARCHAR(10) NOT NULL,
			lang_to VARCHAR(10) NOT NULL,
			used BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}
	_, err = r.db.Pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_ai_phrases_duel_id ON ai_phrases(duel_id)`)
	if err != nil {
		return err
	}
	_, err = r.db.Pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_ai_phrases_used ON ai_phrases(used)`)
	return err
}

// EnsureSchemaFixes applies one-time schema corrections at startup.
// These are idempotent so they are safe to run on every start.
func (r *DuelRepo) EnsureSchemaFixes(ctx context.Context) error {
	// Remove UNIQUE constraint on room_code so multiple duels can share the same
	// room code (rematches). Without this, rematch creation fails and new games
	// reuse old finished duels, corrupting analysis data.
	_, err := r.db.Pool.Exec(ctx, `
		ALTER TABLE duels DROP CONSTRAINT IF EXISTS duels_room_code_key
	`)
	if err != nil {
		return err
	}
	_, err = r.db.Pool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_duels_room_code ON duels(room_code)
	`)
	if err != nil {
		return err
	}
	// Ensure UNIQUE(round_id, participant_id) on player_answers so that
	// ON CONFLICT DO NOTHING works in CreateAnswer.
	// First remove any duplicate rows (keep the best answer per round+participant),
	// then create the unique index (idempotent via IF NOT EXISTS).
	_, err = r.db.Pool.Exec(ctx, `
		DELETE FROM player_answers WHERE ctid IN (
			SELECT ctid FROM (
				SELECT ctid,
					ROW_NUMBER() OVER (
						PARTITION BY round_id, participant_id
						ORDER BY CASE WHEN translation_text != '' THEN 0 ELSE 1 END, ctid
					) AS rn
				FROM player_answers
			) ranked
			WHERE rn > 1
		)
	`)
	if err != nil {
		return err
	}
	_, err = r.db.Pool.Exec(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_player_answers_round_participant
		ON player_answers(round_id, participant_id)
	`)
	if err != nil {
		return err
	}
	// Add valid_answers column to game_rounds for storing all accepted answer variants
	_, err = r.db.Pool.Exec(ctx, `
		ALTER TABLE game_rounds ADD COLUMN IF NOT EXISTS valid_answers TEXT[] DEFAULT '{}'
	`)
	if err != nil {
		return err
	}
	// Expand short VARCHAR columns that break on long AI-generated prompts and custom topics
	for _, ddl := range []string{
		`ALTER TABLE ai_phrases ALTER COLUMN prompt TYPE TEXT`,
		`ALTER TABLE ai_phrases ALTER COLUMN topic TYPE TEXT`,
		`ALTER TABLE phrases     ALTER COLUMN topic TYPE TEXT`,
		`ALTER TABLE duels       ALTER COLUMN theme TYPE TEXT`,
	} {
		if _, err = r.db.Pool.Exec(ctx, ddl); err != nil {
			return err
		}
	}
	return nil
}

// DuelDetail represents detailed duel information for analysis
type DuelDetail struct {
	DuelID       string              `json:"duel_id"`
	RoomCode     string              `json:"room_code"`
	Status       string              `json:"status"`
	WinnerID     string              `json:"winner_user_id"`
	Theme        string              `json:"theme"`
	Difficulty   string              `json:"difficulty"`
	LangFrom     string              `json:"lang_from"`
	LangTo       string              `json:"lang_to"`
	StartedAt    string              `json:"started_at"`
	FinishedAt   string              `json:"finished_at"`
	Participants []ParticipantDetail `json:"participants"`
	Rounds       []RoundDetail       `json:"rounds"`
}

// ParticipantDetail represents player stats in a duel
type ParticipantDetail struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Avatar       string `json:"avatar"`
	FinalHP      int    `json:"final_hp"`
	CorrectCount int    `json:"correct_count"`
	WrongCount   int    `json:"wrong_count"`
	TotalDamage  int    `json:"total_damage"`
	IsWinner     bool   `json:"is_winner"`
}

// RoundDetail represents a single round with answers
type RoundDetail struct {
	RoundNumber int            `json:"round_number"`
	Phrase      string         `json:"phrase"`
	LangFrom    string         `json:"lang_from"`
	LangTo      string         `json:"lang_to"`
	Answers     []AnswerDetail `json:"answers"`
}

// AnswerDetail represents a player's answer in a round
type AnswerDetail struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Answer       string `json:"answer"`
	IsCorrect    bool   `json:"is_correct"`
	ResponseTime int    `json:"response_time_ms"`
	DamageDealt  int    `json:"damage_dealt"`
}

// GetDuelDetails returns full duel analysis with rounds and answers
func (r *DuelRepo) GetDuelDetails(ctx context.Context, duelID, userID string) (*DuelDetail, error) {
	// Get duel info
	log.Printf("GetDuelDetails: duelID=%s userID=%s", duelID, userID)
	row := r.db.Pool.QueryRow(ctx, `
		SELECT d.duel_id, d.room_code, d.status, COALESCE(d.winner_user_id, ''), d.theme, d.difficulty,
			   d.language_from, d.language_to, d.started_at, d.finished_at
		FROM duels d
		WHERE d.duel_id = $1`, duelID)

	var detail DuelDetail
	var startedAt, finishedAt *time.Time
	if err := row.Scan(&detail.DuelID, &detail.RoomCode, &detail.Status, &detail.WinnerID,
		&detail.Theme, &detail.Difficulty, &detail.LangFrom, &detail.LangTo, &startedAt, &finishedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if startedAt != nil {
		detail.StartedAt = startedAt.Format(time.RFC3339)
	}
	if finishedAt != nil {
		detail.FinishedAt = finishedAt.Format(time.RFC3339)
	}

	// Get participants with stats
	rows, err := r.db.Pool.Query(ctx, `
		SELECT dp.user_id, COALESCE(u.username, ''), COALESCE(u.avatar, 'default'), 
			   COALESCE(dp.final_hp, 0)
		FROM duel_participants dp
		JOIN users u ON u.user_id = dp.user_id
		WHERE dp.duel_id = $1`, duelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	detail.Participants = []ParticipantDetail{}
	for rows.Next() {
		var p ParticipantDetail
		if err := rows.Scan(&p.UserID, &p.Username, &p.Avatar, &p.FinalHP); err != nil {
			return nil, err
		}
		if p.UserID == "" {
			continue
		}
		p.IsWinner = p.UserID == detail.WinnerID
		detail.Participants = append(detail.Participants, p)
	}

	// Get rounds with phrases
	roundRows, err := r.db.Pool.Query(ctx, `
		SELECT COALESCE(gr.round_id, ''::uuid)::text, COALESCE(gr.round_number, 0), COALESCE(p.text, '')
		FROM game_rounds gr
		LEFT JOIN phrases p ON p.phrase_id = gr.phrase_id
		WHERE gr.duel_id = $1
		ORDER BY gr.round_number`, duelID)
	if err != nil {
		return nil, err
	}
	defer roundRows.Close()

	roundMap := make(map[string]RoundDetail)
	for roundRows.Next() {
		var roundID string
		var rd RoundDetail
		if err := roundRows.Scan(&roundID, &rd.RoundNumber, &rd.Phrase); err != nil {
			return nil, err
		}
		if roundID == "" {
			continue
		}
		rd.LangFrom = detail.LangFrom
		rd.LangTo = detail.LangTo
		rd.Answers = []AnswerDetail{}
		roundMap[roundID] = rd
	}

	// Get answers for all rounds
	for roundID, rd := range roundMap {
		ansRows, err := r.db.Pool.Query(ctx, `
			SELECT dp.user_id, COALESCE(u.username, ''), COALESCE(pa.translation_text, ''), pa.is_correct,
				   COALESCE(pa.response_time_ms, 0), COALESCE(pa.damage_dealt, 0)
			FROM player_answers pa
			JOIN duel_participants dp ON dp.participant_id = pa.participant_id
			JOIN users u ON u.user_id = dp.user_id
			WHERE pa.round_id = $1
			ORDER BY pa.response_time_ms`, roundID)
		if err != nil {
			return nil, err
		}

		for ansRows.Next() {
			var ad AnswerDetail
			if err := ansRows.Scan(&ad.UserID, &ad.Username, &ad.Answer, &ad.IsCorrect,
				&ad.ResponseTime, &ad.DamageDealt); err != nil {
				ansRows.Close()
				return nil, err
			}
			if ad.UserID == "" {
				continue
			}
			rd.Answers = append(rd.Answers, ad)
			roundMap[roundID] = rd

			// Update participant stats
			for i := range detail.Participants {
				if detail.Participants[i].UserID == ad.UserID {
					if ad.IsCorrect {
						detail.Participants[i].CorrectCount++
					} else {
						detail.Participants[i].WrongCount++
					}
					detail.Participants[i].TotalDamage += ad.DamageDealt
				}
			}
		}
		ansRows.Close()
	}

	detail.Rounds = make([]RoundDetail, 0, len(roundMap))
	for _, rd := range roundMap {
		detail.Rounds = append(detail.Rounds, rd)
	}

	return &detail, nil
}

// GetDuelByID returns duel by ID
func (r *DuelRepo) GetDuelByID(ctx context.Context, duelID string) (*Duel, error) {
	row := r.db.Pool.QueryRow(ctx,
		`SELECT duel_id, room_code, theme, difficulty, language_from, language_to, status
         FROM duels WHERE duel_id = $1`,
		duelID,
	)
	var d Duel
	if err := row.Scan(&d.ID, &d.RoomCode, &d.Theme, &d.Difficulty, &d.LangFrom, &d.LangTo, &d.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

// DuelAnalysis represents a quick duel analysis for display
type DuelAnalysis struct {
	DuelID       string               `json:"duel_id"`
	Participants []ParticipantSummary `json:"participants"`
	Rounds       []RoundAnalysis      `json:"rounds"`
}

type ParticipantSummary struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Correct  int    `json:"correct"`
	Wrong    int    `json:"wrong"`
}

type RoundAnalysis struct {
	RoundNumber   int             `json:"round_number"`
	Phrase        string          `json:"phrase"`
	CorrectAnswer string          `json:"correct_answer"`
	ValidAnswers  []string        `json:"valid_answers"`
	Answers       []AnswerSummary `json:"answers"`
}

type AnswerSummary struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Answer    string `json:"answer"`
	IsCorrect bool   `json:"is_correct"`
}

// GetDuelAnalysis returns a simplified analysis for display after game
func (r *DuelRepo) GetDuelAnalysis(ctx context.Context, duelID string) (*DuelAnalysis, error) {
	analysis := &DuelAnalysis{
		DuelID:       duelID,
		Participants: []ParticipantSummary{},
		Rounds:       []RoundAnalysis{},
	}

	pRows, err := r.db.Pool.Query(ctx, `
		SELECT dp.user_id, COALESCE(u.username, '')
		FROM duel_participants dp
		JOIN users u ON u.user_id = dp.user_id
		WHERE dp.duel_id = $1`, duelID)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()

	participantMap := make(map[string]string)
	for pRows.Next() {
		var uid, username string
		if err := pRows.Scan(&uid, &username); err != nil {
			continue
		}
		analysis.Participants = append(analysis.Participants, ParticipantSummary{UserID: uid, Username: username})
		participantMap[uid] = username
	}

	// Get rounds with phrases
	rRows, err := r.db.Pool.Query(ctx, `
		SELECT gr.round_id::text, COALESCE(gr.round_number, 0), COALESCE(p.text, ''), COALESCE(gr.correct_answer, ''), COALESCE(gr.valid_answers, '{}')
		FROM game_rounds gr
		LEFT JOIN phrases p ON p.phrase_id = gr.phrase_id
		WHERE gr.duel_id = $1
		ORDER BY gr.round_number`, duelID)
	if err != nil {
		return nil, err
	}
	defer rRows.Close()

	roundMap := make(map[string]RoundAnalysis)
	for rRows.Next() {
		var roundID string
		var roundNum int
		var phrase, correctAnswer string
		var validAnswers []string
		if err := rRows.Scan(&roundID, &roundNum, &phrase, &correctAnswer, &validAnswers); err != nil {
			continue
		}
		if roundID == "" {
			continue
		}
		if validAnswers == nil {
			validAnswers = []string{}
		}
		roundMap[roundID] = RoundAnalysis{
			RoundNumber:   roundNum,
			Phrase:        phrase,
			CorrectAnswer: correctAnswer,
			ValidAnswers:  validAnswers,
			Answers:       []AnswerSummary{},
		}
	}

	// Get answers
	for roundID := range roundMap {
		aRows, err := r.db.Pool.Query(ctx, `
			SELECT dp.user_id, COALESCE(u.username, ''), 
				   COALESCE(pa.translation_text, ''), pa.is_correct
			FROM player_answers pa
			JOIN duel_participants dp ON dp.participant_id = pa.participant_id
			JOIN users u ON u.user_id = dp.user_id
			WHERE pa.round_id = $1
			ORDER BY pa.response_time_ms`, roundID)
		if err != nil {
			continue
		}

		for aRows.Next() {
			var uid, username, answer string
			var isCorrect bool
			if err := aRows.Scan(&uid, &username, &answer, &isCorrect); err != nil {
				continue
			}
			if uid == "" {
				continue
			}
			ra := roundMap[roundID]
			ra.Answers = append(ra.Answers, AnswerSummary{
				UserID:    uid,
				Username:  username,
				Answer:    answer,
				IsCorrect: isCorrect,
			})
			roundMap[roundID] = ra

			// Update participant stats
			for i := range analysis.Participants {
				if analysis.Participants[i].UserID == uid {
					if isCorrect {
						analysis.Participants[i].Correct++
					} else {
						analysis.Participants[i].Wrong++
					}
				}
			}
		}
		aRows.Close()
	}

	for _, ra := range roundMap {
		analysis.Rounds = append(analysis.Rounds, ra)
	}

	sort.Slice(analysis.Rounds, func(i, j int) bool {
		return analysis.Rounds[i].RoundNumber < analysis.Rounds[j].RoundNumber
	})

	return analysis, nil
}
