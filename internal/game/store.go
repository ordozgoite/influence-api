package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Store struct {
	redis *redis.Client
}

func NewStore(redisClient *redis.Client) *Store {
	return &Store{
		redis: redisClient,
	}
}

func (store *Store) GetRedis() *redis.Client {
	return store.redis
}

func (store *Store) StartGame(gameID string, sessionToken string) (*PublicGameState, error) {
	ctx := context.Background()

	session, err := store.resolveSession(ctx, gameID, sessionToken)
	if err != nil {
		return nil, err
	}

	playerID := session.PlayerID

	game, err := store.withGameLock(ctx, gameID, func(game *Game) error {
		if game.Started {
			return ErrAlreadyStarted
		}
		if game.Finished {
			return ErrGameAlreadyFinished
		}
		if game.AdminID != playerID {
			return ErrOnlyAdminCanStartGame
		}

		if err := SetupNewGame(game); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	BroadcastEvent(
		ProjectPublicGameState(game),
		"game_started",
		nil,
	)

	return ProjectPublicGameState(game), nil
}

func (store *Store) GetPlayerInfluences(
	gameID string,
	sessionToken string,
) ([]Influence, error) {
	ctx := context.Background()

	session, err := store.resolveSession(ctx, gameID, sessionToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to resolve session.")
		return nil, err
	}

	actingPlayerID := session.PlayerID
	gameKey := "game:" + gameID

	gameJSON, err := store.redis.Get(ctx, gameKey).Bytes()
	if err == redis.Nil {
		return nil, ErrGameNotFound
	}
	if err != nil {
		return nil, err
	}

	var game Game
	if err := json.Unmarshal(gameJSON, &game); err != nil {
		return nil, err
	}

	if !game.Started || game.Finished {
		return nil, ErrNotStarted
	}

	for _, player := range game.Players {
		if player.ID == actingPlayerID {
			return player.Influences, nil
		}
	}

	return nil, ErrPlayerNotFound
}

func getTurnPlayer(game *Game, actingPlayerID string) (*Player, error) {
	turnPlayer := game.Players[game.TurnIndex]
	if turnPlayer.ID != actingPlayerID {
		return nil, fmt.Errorf("not_your_turn")
	}
	return turnPlayer, nil
}

func findPlayerByID(game *Game, playerID string) (*Player, error) {
	for _, p := range game.Players {
		if p.ID == playerID {
			return p, nil
		}
	}
	return nil, ErrPlayerNotFound
}

func advanceTurn(game *Game) {
	game.TurnIndex = (game.TurnIndex + 1) % len(game.Players)
}

func validateActionContext(
	game *Game,
	actingPlayerID string,
) (*Player, error) {

	if !game.Started || game.Finished {
		return nil, ErrNotStarted
	}

	return getTurnPlayer(game, actingPlayerID)
}

func (store *Store) DeclareAction(
	gameID string,
	action DeclareActionPayload,
	sessionToken string,
) (*PublicGameState, error) {

	ctx := context.Background()

	session, err := store.resolveSession(ctx, gameID, sessionToken)
	if err != nil {
		return nil, err
	}

	actingPlayerID := session.PlayerID

	var resultGame *Game
	var actionPayload DeclareActionPayload

	resultGame, err = store.withGameLock(ctx, gameID, func(game *Game) error {

		turnPlayer, err := validateActionContext(game, actingPlayerID)
		if err != nil {
			return err
		}

		actionPayload, err = applyAction(game, turnPlayer, action)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	BroadcastEvent(
		ProjectPublicGameState(resultGame),
		"action_declared",
		map[string]any{
			"actionPayload": actionPayload,
		},
	)

	return ProjectPublicGameState(resultGame), nil
}

func applyAction(
	game *Game,
	actor *Player,
	action DeclareActionPayload,
) (DeclareActionPayload, error) {

	switch action.ActionName {
	case "income":
		return declareIncome(game, actor), nil

	case "foreign_aid":
		return declareForeignAid(actor), nil

	case "coup":
		if action.TargetPlayerID == nil {
			return DeclareActionPayload{}, ErrPlayerNotFound
		}

		target, err := findPlayerByID(game, *action.TargetPlayerID)
		if err != nil {
			return DeclareActionPayload{}, err
		}

		return declareCoup(game, actor, target)

	default:
		return DeclareActionPayload{}, errors.New("invalid_action_name")
	}
}

func declareIncome(game *Game, player *Player) DeclareActionPayload {
	player.Coins++
	advanceTurn(game)

	return DeclareActionPayload{
		ID:                  uuid.New().String(),
		ActionName:          "income",
		ActorPlayerID:       player.ID,
		ActorPlayerNickname: player.Nickname,
		RequiresTarget:      false,
		IsImmediate:         true,
		BlockableRoles:      []string{},
		IsContestable:       false,
	}
}

func declareForeignAid(player *Player) DeclareActionPayload {
	return DeclareActionPayload{
		ID:                  uuid.New().String(),
		ActionName:          "foreign_aid",
		ActorPlayerID:       player.ID,
		ActorPlayerNickname: player.Nickname,
		RequiresTarget:      false,
		IsImmediate:         false,
		BlockableRoles:      []string{"Duke"},
		IsContestable:       false,
	}
}

func declareCoup(game *Game, actor *Player, target *Player) (DeclareActionPayload, error) {
	if actor.Coins < 7 {
		return DeclareActionPayload{}, fmt.Errorf("not_enough_coins")
	}

	if !target.Alive {
		return DeclareActionPayload{}, fmt.Errorf("target_player_is_dead")
	}

	actor.Coins -= 7
	advanceTurn(game)

	return DeclareActionPayload{
		ID:                   uuid.New().String(),
		ActionName:           "coup",
		ActorPlayerID:        actor.ID,
		ActorPlayerNickname:  actor.Nickname,
		RequiresTarget:       true,
		TargetPlayerID:       &target.ID,
		TargetPlayerNickname: &target.Nickname,
		IsImmediate:          true,
		BlockableRoles:       []string{},
		IsContestable:        false,
	}, nil
}

// func (store *Store) BlockAction(
// 	gameID string,
// 	actionID string,
// 	blockingRole string,
// 	sessionToken string,
// ) (*PublicGameState, error) {
// 	ctx := context.Background()

// 	session, err := store.resolveSession(ctx, gameID, sessionToken)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to resolve session.")
// 		return nil, err
// 	}

// 	blockingPlayerID := session.PlayerID
// 	gameKey := "game:" + gameID

// 	var game Game
// 	if err := json.Unmarshal(gameJSON, &game); err != nil {
// 		return nil, err
// 	}

// 	if !game.Started || game.Finished {
// 		return nil, ErrNotStarted
// 	}

// 	// TODO:
// 	// Procurar PendingAction com o actionID
// 	// Se o PendingAction não existir, retornar erro
// 	// Se o PendingAction existir, verificar se o blockingRole é válido
// 	// Se o blockingRole for válido, criar pending action de bloqueio
// 	// Se o blockingRole for inválido, retornar erro
// 	// Atualizar o game com o novo PendingAction
// 	// Retornar o novo game

// 	return game.GetPublicGameState(), nil
// }
