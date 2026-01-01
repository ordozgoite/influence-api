package game

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func (store *Store) resolveSession(
	ctx context.Context,
	gameID string,
	sessionToken string,
) (*PlayerSession, error) {
	if sessionToken == "" {
		return nil, ErrInvalidSession
	}

	data, err := store.redis.Get(ctx, "session:"+sessionToken).Bytes()
	if err == redis.Nil {
		return nil, ErrInvalidSession
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get session from Redis.")
		return nil, err
	}

	var session PlayerSession
	if err := json.Unmarshal(data, &session); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal session.")
		return nil, err
	}

	if session.GameID != gameID {
		return nil, ErrInvalidSession
	}

	return &session, nil
}

func (store *Store) CreatePlayerSession(gameID string, playerID string) (string, error) {
	if store.redis == nil {
		return "", errors.New("redis_not_configured")
	}

	ctx := context.Background()

	sessionToken := uuid.NewString()

	session := PlayerSession{
		PlayerID: playerID,
		GameID:   gameID,
	}

	data, err := json.Marshal(session)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize session.")
		return "", err
	}

	redisKey := "session:" + sessionToken

	err = store.redis.Set(ctx, redisKey, data, SessionDuration).Err()
	if err != nil {
		log.Error().Err(err).Msg("Failed to save session to Redis.")
		return "", err
	}

	return sessionToken, nil
}
