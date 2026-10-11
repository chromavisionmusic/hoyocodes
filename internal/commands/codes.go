package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromavisionmusic/hoyocodes/internal/config"
	"github.com/chromavisionmusic/hoyocodes/internal/hoyo"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var CodesCommandMetadata = discord.SlashCommandCreate{
	Name:        "codes",
	Description: "Get active redeem codes for HoYoverse games",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "game",
			Description: "The game to get active codes for",
			Required:    true,
			Choices: []discord.ApplicationCommandOptionChoiceString{
				{
					Name:  "Genshin Impact",
					Value: "genshin",
				},
				{
					Name:  "Honkai: Star Rail",
					Value: "hkrpg",
				},
				{
					Name:  "Zenless Zone Zero",
					Value: "nap",
				},
			},
		},
	},
}

func CodesCommandHandler(cfg *config.Config) handler.SlashCommandHandler {
	return func(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
		// Defer response publicly so everyone in the channel can view the codes
		if err := e.DeferCreateMessage(false); err != nil {
			return err
		}

		game := data.String("game")
		gameInfo, ok := hoyo.GetGameInfo(game)
		if !ok {
			_, err := e.CreateFollowupMessage(discord.MessageCreate{
				Content: "Invalid game selection.",
			})
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		codes, err := hoyo.DefaultClient.GetCodes(ctx, game)
		if err != nil {
			embed := discord.NewEmbed().
				WithTitle(fmt.Sprintf("%s — Codes Unavailable", gameInfo.Name)).
				WithColor(0xED4245).
				WithDescription("Failed to retrieve active codes at this time. Please try again later.").
				WithTimestamp(time.Now())

			_, fErr := e.CreateFollowupMessage(discord.MessageCreate{
				Embeds: []discord.Embed{embed},
			})
			return fErr
		}

		embed := discord.NewEmbed().
			WithTitle(fmt.Sprintf("%s | Active Codes", gameInfo.Name)).
			WithColor(gameInfo.Color).
			WithFooter(cfg.Bot.Name, "").
			WithTimestamp(time.Now())

		if gameInfo.Thumbnail != "" {
			embed = embed.WithThumbnail(gameInfo.Thumbnail)
		}

		embed = embed.WithDescription(BuildCodesDescription(codes, gameInfo))

		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Embeds: []discord.Embed{embed},
		})
		return err
	}
}

func BuildCodesDescription(codes []hoyo.CodeItem, gameInfo hoyo.GameInfo) string {
	if len(codes) == 0 {
		return "There are no active codes"
	}

	var sb strings.Builder
	for i, c := range codes {
		var formattedRewards string
		redeemURL := gameInfo.FormatRedeemURL(c.Code)
		if c.Rewards != "" {
			formattedRewards = strings.ReplaceAll(strings.ReplaceAll(c.Rewards, ";", "\n> - "), "*", " x")
		} else {
			formattedRewards = "Unknown"
		}
		sb.WriteString(fmt.Sprintf("**%d. [%s](%s)**Rewards\n> - %s\n\n", i+1, c.Code, redeemURL, formattedRewards))
	}

	desc := strings.TrimSpace(sb.String())
	if len(desc) > 4096 {
		desc = desc[:4093] + "..."
	}
	return desc
}
