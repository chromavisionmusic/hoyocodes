package commands

import (
	"github.com/chromavisionmusic/hoyocodes/internal/config"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var Commands = []discord.ApplicationCommandCreate{
	PingCommandMetadata,
	AboutCommandMetadata,
}

func LoadCommands(handler *handler.Mux, cfg *config.Config) {
	handler.SlashCommand("/ping", PingCommandHandler)
	handler.SlashCommand("/about", AboutCommandHandler(cfg))
}
