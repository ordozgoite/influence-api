package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"influence_game/actions"
)

func main() {

	// ─────────────────────────────────────────────────────────────
	// 👇 Habilita logs coloridos no modo desenvolvimento
	// ─────────────────────────────────────────────────────────────
	if os.Getenv("GO_ENV") != "production" {
		writer := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		log.Logger = log.Output(writer)
	}
	// ─────────────────────────────────────────────────────────────

	app := actions.App()

	if err := app.Serve(); err != nil {
		log.Fatal().Err(err).Msg("server failed to start")
	}
}
