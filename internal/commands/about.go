package commands

import (
	"fmt"
	"time"

	"github.com/chromavisionmusic/hoyocodes/internal/config"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var startTime = time.Now()

var AboutCommandMetadata = discord.SlashCommandCreate{
	Name:        "about",
	Description: "Information about HoyoCodes",
}

func AboutCommandHandler(cfg *config.Config) handler.SlashCommandHandler {
	return func(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
		embed := discord.NewEmbed().
			WithTitle(cfg.Bot.Name).
			WithColor(0x5865F2).
			WithDescription(cfg.Bot.Description).
			WithFooter("Built with disgo", "https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png").
			WithTimestamp(time.Now())

		if selfUser, ok := e.Client().Caches.SelfUser(); ok {
			embed = embed.WithThumbnail(selfUser.EffectiveAvatarURL())
		}

		users := 0
		guilds := 0
		if e.Client().Caches != nil {
			guilds = e.Client().Caches.GuildsLen()
			for guild := range e.Client().Caches.Guilds() {
				users += guild.MemberCount
			}
			if users == 0 {
				users = e.Client().Caches.MembersAllLen()
			}
		}

		ping := int64(0)
		if e.Client().HasGateway() && e.Client().Gateway != nil {
			ping = e.Client().Gateway.Latency().Milliseconds()
		}

		embed = embed.
			AddField("Version", cfg.Bot.Version, true).
			AddField("Ping", fmt.Sprintf("%dms", ping), true).
			AddField("Uptime", fmt.Sprintf("<t:%d:R>", startTime.Unix()), true).
			AddField("Users", fmt.Sprintf("%d", users), true).
			AddField("Guilds", fmt.Sprintf("%d", guilds), true).
			AddField("Commands", fmt.Sprintf("%d", len(Commands)), true)

		return e.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{embed},
		})
	}
}
