package rooms

import "github.com/gobuffalo/buffalo"

func Register(app *buffalo.App, controller *RoomsController) {
	app.POST("/rooms", controller.CreateRoom)
	app.POST("/rooms/{joinCode}/join", controller.JoinRoom)
	app.POST("/rooms/{gameID}/start", controller.StartGame)

	// In-game routes
	app.GET("/games/{gameID}/player/influences", controller.GetPlayerInfluences)
	app.POST("/games/{gameID}/actions/declare", controller.DeclareAction)
}
