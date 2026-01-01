package game

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func (store *Store) withGameLock(
	ctx context.Context,
	gameID string,
	fn func(*Game) error,
) (*Game, error) {

	gameKey := "game:" + gameID
	var updatedGame *Game

	for {
		err := store.redis.Watch(ctx, func(tx *redis.Tx) error {
			gameJSON, err := tx.Get(ctx, gameKey).Bytes()
			if err == redis.Nil {
				return ErrGameNotFound
			}
			if err != nil {
				log.Error().Err(err).Msg("Failed to get game from Redis.")
				return err
			}

			var game Game
			if err := json.Unmarshal(gameJSON, &game); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal game.")
				return err
			}

			if err := fn(&game); err != nil {
				return err
			}

			updatedJSON, err := json.Marshal(&game)
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshal game.")
				return err
			}

			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, gameKey, updatedJSON, 0)
				return nil
			})

			if err == nil {
				updatedGame = &game
			}

			return err
		}, gameKey)

		if err == redis.TxFailedErr {
			continue
		}

		if err != nil {
			return nil, err
		}

		break
	}

	return updatedGame, nil
}
