package commands

import (
	"github.com/chromavisionmusic/hoyocodes/internal/config"
	"github.com/chromavisionmusic/hoyocodes/internal/database"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var Commands = []discord.ApplicationCommandCreate{
	PingCommandMetadata,
	AboutCommandMetadata,
	CodesCommandMetadata,
}

func LoadCommands(handler *handler.Mux, cfg *config.Config, db ...*database.DB) {
	handler.SlashCommand("/ping", PingCommandHandler)
	handler.SlashCommand("/about", AboutCommandHandler(cfg))
	handler.SlashCommand("/codes", CodesCommandHandler(cfg))
}

