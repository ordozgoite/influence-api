package game

import "time"

type ActionType struct {
	name            string
	isImmediate     bool
	isBlockable     bool
	isContestable   bool
	requiresTarget  bool
	bloackableRoles []Influence
	targetPlayerID  *string
}

type DeclareActionPayload struct {
	ID                   string   `json:"id"`
	ActionName           string   `json:"actionName"`
	ActorPlayerID        string   `json:"actorPlayerId"`
	ActorPlayerNickname  string   `json:"actorPlayerNickname"`
	RequiresTarget       bool     `json:"requiresTarget"`
	TargetPlayerID       *string  `json:"targetPlayerId,omitempty"`
	TargetPlayerNickname *string  `json:"targetPlayerNickname,omitempty"`
	IsImmediate          bool     `json:"isImmediate"`
	BlockableRoles       []string `json:"blockableRoles"`
	IsContestable        bool     `json:"isContestable"`
}

type Influence struct {
	Role     string `json:"role"`
	Revealed bool   `json:"revealed"`
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
	AdminID   string
	JoinCode  string
	Players   []*Player
	TurnIndex int
	Started   bool
	Finished  bool

	Deck []Influence `json:"deck"`
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

type PlayerPublicInfo struct {
	ID         string            `json:"id"`
	Nickname   string            `json:"nickname"`
	Coins      int               `json:"coins"`
	Alive      bool              `json:"alive"`
	Influences []PublicInfluence `json:"influences"`
}

type PublicGameState struct {
	GameID     string             `json:"gameID"`
	JoinCode   string             `json:"joinCode"`
	Started    bool               `json:"started"`
	AdminID    string             `json:"adminID"`
	Finished   bool               `json:"finished"`
	TurnIndex  int                `json:"turnIndex"`
	Players    []PlayerPublicInfo `json:"players"`
	DeckLength int                `json:"deckLength"`
}

type PendingAction struct {
	ID        string     `json:"id"`
	ActorID   string     `json:"actorId"`
	Action    ActionType `json:"action"`
	TargetID  *string    `json:"targetId,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	Status    string     `json:"status"` // "declared", "resolved", "canceled"...
}

type OnboardingResult struct {
	Game   *PublicGameState `json:"game"`
	Player *Player          `json:"player"`
	Token  string           `json:"token"`
}
