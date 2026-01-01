package game

func ProjectPublicGameState(game *Game) *PublicGameState {
	playersPublicInfo := make([]PlayerPublicInfo, 0, len(game.Players))

	for _, player := range game.Players {
		playersPublicInfo = append(
			playersPublicInfo,
			projectPublicPlayerInfo(player),
		)
	}

	return &PublicGameState{
		GameID:     game.ID,
		JoinCode:   game.JoinCode,
		Started:    game.Started,
		Finished:   game.Finished,
		TurnIndex:  game.TurnIndex,
		Players:    playersPublicInfo,
		AdminID:    game.AdminID,
		DeckLength: len(game.Deck),
	}
}

func projectPublicPlayerInfo(player *Player) PlayerPublicInfo {
	influences := make([]PublicInfluence, 0, len(player.Influences))

	for _, influence := range player.Influences {
		if influence.Revealed {
			influences = append(influences, PublicInfluence{
				Role:     &influence.Role,
				Revealed: influence.Revealed,
			})
		} else {
			influences = append(influences, PublicInfluence{
				Role:     nil,
				Revealed: influence.Revealed,
			})
		}
	}

	return PlayerPublicInfo{
		ID:         player.ID,
		Nickname:   player.Nickname,
		Coins:      player.Coins,
		Alive:      player.Alive,
		Influences: influences,
	}
}
