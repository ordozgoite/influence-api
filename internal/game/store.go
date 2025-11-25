package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type ActionType string

const (
	ActionStart   ActionType = "start"
	ActionEndTurn ActionType = "end_turn"

	ActionIncome     ActionType = "income"
	ActionForeignAid ActionType = "foreign_aid"
	ActionCoup       ActionType = "coup"

	ActionTax         ActionType = "tax"
	ActionAssassinate ActionType = "assassinate"
	ActionSteal       ActionType = "steal"
	ActionExchange    ActionType = "exchange"

	ActionBlockForeignAid  ActionType = "block_foreign_aid"
	ActionBlockAssassinate ActionType = "block_assassinate"
	ActionBlockSteal       ActionType = "block_steal"
)

var (
	ErrGameNotFound   = errors.New("game_not_found")
	ErrAlreadyStarted = errors.New("game_already_started")
	ErrNotStarted     = errors.New("game_not_started")
	ErrInvalidAction  = errors.New("invalid_action")
)

type Influence struct {
	Role     string       `json:"role"`
	Revealed bool         `json:"revealed"`
	Actions  []ActionType `json:"actions"`
}

type Player struct {
	ID         string      `json:"id"`
	Nickname   string      `json:"nickname"`
	Coins      int         `json:"coins"`
	Alive      bool        `json:"alive"`
	Influences []Influence `json:"influences"`
}

type Game struct {
	ID        string
	CreatedAt time.Time
	Players   []*Player
	TurnIndex int
	Started   bool
	Finished  bool
}

type PlayerSession struct {
	PlayerID string `json:"playerId"`
	GameID   string `json:"gameId"`
}

/*
⚠️ Warning:
- this is a public representation of the influence, so it should not contain the role if it is not revealed
*/
type PublicInfluence struct {
	Role     *string `json:"role,omitempty"`
	Revealed bool    `json:"revealed"`
}

type PublicPlayer struct {
	ID         string            `json:"id"`
	Nickname   string            `json:"nickname"`
	Coins      int               `json:"coins"`
	Alive      bool              `json:"alive"`
	Influences []PublicInfluence `json:"influences"`
}

type PublicState struct {
	GameID    string         `json:"gameID"`
	Started   bool           `json:"started"`
	Finished  bool           `json:"finished"`
	TurnIndex int            `json:"turnIndex"`
	Players   []PublicPlayer `json:"players"`
}

func newGame() *Game {
	return &Game{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		Players:   []*Player{},
	}
}

func (game *Game) getPublicState() *PublicState {
	publicPlayers := make([]PublicPlayer, 0, len(game.Players))

	for _, player := range game.Players {
		publicPlayers = append(publicPlayers, PublicPlayer{
			ID: player.ID, Nickname: player.Nickname, Coins: player.Coins, Alive: player.Alive,
		})
	}

	return &PublicState{
		GameID: game.ID, Started: game.Started, Finished: game.Finished,
		TurnIndex: game.TurnIndex, Players: publicPlayers,
	}
}

func (g *Game) handleAction(action ActionType, body json.RawMessage) error {
	switch action {
	case ActionStart:
		if g.Started {
			return ErrAlreadyStarted
		}
		if len(g.Players) < 2 {
			return errors.New("need_at_least_two_players")
		}
		g.Started = true
		for _, p := range g.Players {
			p.Coins = 2
			p.Alive = true
		}
		return nil
	case ActionIncome:
		if !g.Started {
			return ErrNotStarted
		}
		cur := g.Players[g.TurnIndex%len(g.Players)]
		if !cur.Alive {
			g.TurnIndex = (g.TurnIndex + 1) % len(g.Players)
			return nil
		}
		cur.Coins++
		g.TurnIndex = (g.TurnIndex + 1) % len(g.Players)
		return nil

		// TODO: Implement other actions
	default:
		return ErrInvalidAction
	}
}

type Store struct {
	mu    sync.RWMutex
	games map[string]*Game

	redis *redis.Client
}

func NewStore(redisClient *redis.Client) *Store {
	return &Store{
		games: make(map[string]*Game),
		redis: redisClient,
	}
}

func (store *Store) NewGame() (*Game, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	newGame := newGame()

	store.games[newGame.ID] = newGame

	serializedGame, err := json.Marshal(newGame)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize game.")
		return nil, err
	}

	redisKey := "game:" + newGame.ID
	ctx := context.Background()

	if err := store.redis.Set(ctx, redisKey, serializedGame, 0).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to save game to Redis.")
		return nil, err
	}

	return newGame, nil
}

func (store *Store) CreatePlayerSession(ctx context.Context, gameID string, playerID string) (string, error) {
	if store.redis == nil {
		return "", errors.New("redis_not_configured")
	}

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

	// Sessão válida por 24 horas (pode ajustar)
	err = store.redis.Set(ctx, redisKey, data, 24*time.Hour).Err()
	if err != nil {
		log.Error().Err(err).Msg("Failed to save session to Redis.")
		return "", err
	}

	return sessionToken, nil
}

func (s *Store) Get(id string) *Game {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.games[id]
}

func (store *Store) Join(gameID, nickname string) (*Game, *Player, string, error) {
	fmt.Println("JOIN REQUEST gameID:", gameID)
	fmt.Println("AVAILABLE GAMES IN REDIS:")

	keys, _ := store.redis.Keys(ctx, "game:*").Result()
	fmt.Println(keys)
	ctx := context.Background()

	// 1. Buscar jogo no Redis
	redisKey := "game:" + gameID
	gameJSON, err := store.redis.Get(ctx, redisKey).Bytes()
	if err == redis.Nil {
		fmt.Println("Game not found")
		return nil, nil, "", ErrGameNotFound
	}
	if err != nil {
		fmt.Println("Error getting game from Redis:", err)
		return nil, nil, "", err
	}

	// 2. Desserializar
	var gameInstance Game
	if err := json.Unmarshal(gameJSON, &gameInstance); err != nil {
		return nil, nil, "", err
	}

	// 3. Verificar se já começou
	if gameInstance.Started {
		return nil, nil, "", ErrAlreadyStarted
	}

	// 4. Criar player
	newPlayer := &Player{
		ID:       uuid.NewString(),
		Nickname: nickname,
		Coins:    0,
		Alive:    false,
	}
	gameInstance.Players = append(gameInstance.Players, newPlayer)

	// 5. Gerar token de sessão
	sessionToken, err := store.CreatePlayerSession(ctx, gameID, newPlayer.ID)
	if err != nil {
		return nil, nil, "", err
	}

	// 6. Re-salvar o jogo no Redis
	updatedJSON, err := json.Marshal(gameInstance)
	if err != nil {
		return nil, nil, "", err
	}

	if err := store.redis.Set(ctx, redisKey, updatedJSON, 0).Err(); err != nil {
		return nil, nil, "", err
	}

	return &gameInstance, newPlayer, sessionToken, nil
}
