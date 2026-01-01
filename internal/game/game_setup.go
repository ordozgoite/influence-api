package game

import "math/rand"

func SetupNewGame(game *Game) error {
	if len(game.Players) < 2 {
		return ErrNeedAtLeastTwoPlayers
	}
	if len(game.Players) > 6 {
		return ErrTooManyPlayers
	}

	game.Started = true
	game.TurnIndex = rand.Intn(len(game.Players))

	deck := NewBaseDeck()
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	for _, p := range game.Players {
		p.Coins = 2
		p.Alive = true
		p.Influences = []Influence{deck[0], deck[1]}
		deck = deck[2:]
	}

	game.Deck = deck
	return nil
}

func NewBaseDeck() []Influence {
	return []Influence{
		{Role: "Duke"},
		{Role: "Duke"},
		{Role: "Duke"},

		{Role: "Assassin"},
		{Role: "Assassin"},
		{Role: "Assassin"},

		{Role: "Ambassador"},
		{Role: "Ambassador"},
		{Role: "Ambassador"},

		{Role: "Captain"},
		{Role: "Captain"},
		{Role: "Captain"},

		{Role: "Contessa"},
		{Role: "Contessa"},
		{Role: "Contessa"},
	}
}
