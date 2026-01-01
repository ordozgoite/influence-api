package game

import "errors"

var (
	ErrGameNotFound          = errors.New("game_not_found")
	ErrAlreadyStarted        = errors.New("game_already_started")
	ErrNotStarted            = errors.New("game_not_started")
	ErrInvalidAction         = errors.New("invalid_action")
	ErrPlayerAlreadyJoined   = errors.New("Player already joined with this nickname")
	ErrGameAlreadyFinished   = errors.New("game_already_finished")
	ErrOnlyAdminCanStartGame = errors.New("only_admin_can_start_game")
	ErrNeedAtLeastTwoPlayers = errors.New("need_at_least_two_players")
	ErrTooManyPlayers        = errors.New("too_many_players")
	ErrInvalidSession        = errors.New("invalid_session")
	ErrNotEnoughInfluences   = errors.New("not_enough_influences")
	ErrPlayerNotFound        = errors.New("player_not_found")
)
