package game

import (
	"encoding/json"
	"influence_game/internal/realtime"
	"time"

	"github.com/rs/zerolog/log"
)

// type TargetedEvent struct {
// 	PlayerID string
// 	Data     any
// }

type ServerEvent struct {
	EventType string           `json:"eventType"`
	GameID    string           `json:"gameID"`
	Timestamp time.Time        `json:"timestamp"`
	GameState *PublicGameState `json:"state,omitempty"`
	Payload   map[string]any   `json:"payload,omitempty"`
}

func BroadcastEvent(
	state *PublicGameState,
	eventType string,
	payload map[string]any,
) {
	if state == nil {
		return
	}

	ev := ServerEvent{
		EventType: eventType,
		GameID:    state.GameID,
		Timestamp: time.Now().UTC(),
		GameState: state,
		Payload:   payload,
	}

	data, err := json.Marshal(ev)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal event.")
		return
	}

	realtime.Manager.Broadcast(state.GameID, data)
}

func SendToPlayer(
	playerID string,
	eventType string,
	gameID string,
	payload map[string]any,
) {
	ev := ServerEvent{
		EventType: eventType,
		GameID:    gameID,
		Timestamp: time.Now().UTC(),
		GameState: nil,
		Payload:   payload,
	}

	data, err := json.Marshal(ev)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal private event.")
		return
	}

	realtime.Manager.SendToPlayer(gameID, playerID, data)
}

// func SendPrivateEvents(
// 	gameID string,
// 	eventType string,
// 	events []TargetedEvent,
// ) {
// 	for _, e := range events {
// 		payload, ok := e.Data.(map[string]any)
// 		if !ok {
// 			log.Error().Msg("Invalid private payload format")
// 			continue
// 		}

// 		SendToPlayer(e.PlayerID, eventType, gameID, payload)
// 	}
// }
