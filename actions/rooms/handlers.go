package rooms

import (
	"influence_game/internal/game"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
)

var renderer = render.New(render.Options{})

type RoomsController struct {
	Store *game.Store
}

func NewRoomsController(store *game.Store) *RoomsController {
	return &RoomsController{Store: store}
}

func (controller *RoomsController) CreateRoom(ctx buffalo.Context) error {
	newGame := controller.Store.NewGame()

	return ctx.Render(200, renderer.JSON(map[string]any{
		"gameID": newGame.ID,
	}))
}
