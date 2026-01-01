package game

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func (store *Store) CreateGameRoom(adminNickname string) (*OnboardingResult, error) {
	adminPlayer := buildNewPlayer(adminNickname)

	newGame, err := store.buildNewGame(adminPlayer)
	if err != nil {
		return nil, err
	}

	if err := store.saveGameToRedis(newGame); err != nil {
		ctx := context.Background()
		_ = store.redis.Del(ctx, "joincode:"+newGame.JoinCode).Err()
		return nil, err
	}

	sessionToken, err := store.CreatePlayerSession(newGame.ID, adminPlayer.ID)
	if err != nil {
		return nil, err
	}

	publicState := ProjectPublicGameState(newGame)

	return &OnboardingResult{
		Game:   publicState,
		Player: adminPlayer,
		Token:  sessionToken,
	}, nil
}

func buildNewPlayer(nickname string) *Player {
	return &Player{
		ID:         uuid.NewString(),
		Nickname:   nickname,
		Coins:      2,
		Alive:      true,
		Influences: []Influence{},
	}
}

func (store *Store) buildNewGame(adminPlayer *Player) (*Game, error) {
	gameID := uuid.NewString()

	joinCode, err := store.reserveJoinCode(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique join code: %w", err)
	}

	game := &Game{
		ID:        gameID,
		CreatedAt: time.Now(),
		Players:   []*Player{adminPlayer},
		JoinCode:  joinCode,
		AdminID:   adminPlayer.ID,
		TurnIndex: 0,
		Started:   false,
		Finished:  false,
		Deck:      []Influence{},
	}

	return game, nil
}

const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomJoinCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (store *Store) reserveJoinCode(gameID string) (string, error) {
	ctx := context.Background()

	for {
		code := randomJoinCode()
		key := "joincode:" + code

		ok, err := store.redis.SetNX(ctx, key, gameID, JoinCodeTTL).Result()
		if err != nil {
			return "", err
		}

		if ok {
			return code, nil
		}
	}
}

func (store *Store) saveGameToRedis(newGame *Game) error {
	serializedGame, err := json.Marshal(newGame)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize game.")
		return err
	}

	redisKey := "game:" + newGame.ID
	ctx := context.Background()

	if err := store.redis.Set(ctx, redisKey, serializedGame, 0).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to save game to Redis.")
		return err
	}

	return nil
}

func (store *Store) Join(joinCode, nickname string) (*OnboardingResult, error) {
	ctx := context.Background()

	joinKey := "joincode:" + joinCode
	gameID, err := store.redis.Get(ctx, joinKey).Result()
	if err == redis.Nil {
		return nil, ErrGameNotFound
	}
	if err != nil {
		return nil, err
	}

	var joinedPlayer *Player

	game, err := store.withGameLock(ctx, gameID, func(game *Game) error {

		if game.Started {
			return ErrAlreadyStarted
		}

		for _, p := range game.Players {
			if p.Nickname == nickname {
				return ErrPlayerAlreadyJoined
			}
		}

		joinedPlayer = buildNewPlayer(nickname)
		game.Players = append(game.Players, joinedPlayer)

		return nil
	})

	if err != nil {
		return nil, err
	}

	sessionToken, err := store.CreatePlayerSession(game.ID, joinedPlayer.ID)
	if err != nil {
		return nil, err
	}

	BroadcastEvent(
		ProjectPublicGameState(game),
		"player_joined",
		map[string]any{
			"newPlayer": joinedPlayer,
		},
	)

	return &OnboardingResult{
		Game:   ProjectPublicGameState(game),
		Player: joinedPlayer,
		Token:  sessionToken,
	}, nil
}
