package actions

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"influence_game/internal/game"
	"influence_game/internal/realtime"

	"github.com/gobuffalo/buffalo"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// depois dá pra restringir por domínio
		return true
	},
}

func GameWebSocketHandler(c buffalo.Context) error {
	r := c.Request()

	gameID := c.Param("gameID")
	if gameID == "" {
		return c.Error(http.StatusBadRequest, errors.New("missing gameID"))
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		log.Error().Msg("Missing token in query params.")
		return c.Error(http.StatusUnauthorized, errors.New("missing token"))
	}

	ctx := context.Background()
	sessionKey := "session:" + token

	sessionJSON, err := gameStore.GetRedis().Get(ctx, sessionKey).Bytes()
	if err == redis.Nil {
		log.Error().Msg("Session not found for token: " + token)
		return c.Error(http.StatusUnauthorized, errors.New("session not found"))
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get session from Redis.")
		return c.Error(http.StatusUnauthorized, errors.New("failed to get session from Redis"))
	}

	var session game.PlayerSession
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal session.")
		return c.Error(http.StatusUnauthorized, errors.New("failed to decode session"))
	}

	if session.GameID != gameID {
		log.Error().Msg("Invalid session game ID.")
		return c.Error(http.StatusUnauthorized, errors.New("invalid game session"))
	}

	conn, err := wsUpgrader.Upgrade(c.Response(), r, nil)
	if err != nil {
		return err
	}

	client := &realtime.Client{
		Conn:     conn,
		GameID:   gameID,
		PlayerID: session.PlayerID,
	}

	realtime.Manager.AddClient(client)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			realtime.Manager.RemoveClient(client)
			_ = conn.Close()
			return nil
		}
	}
}

func extractBearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
		return ""
	}
	return header[len(prefix):]
}
