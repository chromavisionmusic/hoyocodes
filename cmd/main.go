package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// Import DisGo packages
	"github.com/chromavisionmusic/hoyocodes/internal/commands"
	"github.com/chromavisionmusic/hoyocodes/internal/config"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

func main() {
	var environment string

	if len(os.Args) < 2 {
		log.Fatal("Environment not specified")
	} else {
		environment = os.Args[1]
	}

	env := config.LoadEnv(environment)
	cfg, err := config.Load("config.yml")
	if err != nil {
		log.Fatal(err)
	}

	// Command handler
	cmd := handler.New()

	// Create a new client instance
	client, err := disgo.New(env.Token,
		bot.WithEventListeners(cmd),
		// Set gateway configuration options
		bot.WithGatewayConfigOpts(
			// Set enabled intents
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentDirectMessages,
			),
			// Set presence from config
			gateway.WithPresenceOpts(getPresenceOpts(cfg.Bot.Presence)...),
		),
		// Enable guild caching so guild list and member counts are stored
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds, cache.FlagMembers),
		),
		// Listen to the Ready event in order to know when the bot is connected
		bot.WithEventListenerFunc(func(e *events.Ready) {
			fmt.Println("Bot is connected as", e.User.Username)
		}),
	)
	if err != nil {
		panic(err)
	}

	commands.LoadCommands(cmd, cfg)
	if err := handler.SyncCommands(client, commands.Commands, []snowflake.ID{env.GuildId}); err != nil {
		panic(err)
	}
	fmt.Printf("Synced %d commands\n", len(commands.Commands))

	// Ensure we close the client on exit
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		client.Close(ctx)
	}()

	// Connect to the gateway
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = client.OpenGateway(ctx); err != nil {
		panic(err)
	}

	// Wait here until CTRL+C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	fmt.Println("Shutting down bot...")
}

func getPresenceOpts(presence config.Presence) []gateway.PresenceOpt {
	var opts []gateway.PresenceOpt

	switch strings.ToLower(presence.Status) {
	case "dnd", "do_not_disturb":
		opts = append(opts, gateway.WithOnlineStatus(discord.OnlineStatusDND))
	case "idle":
		opts = append(opts, gateway.WithOnlineStatus(discord.OnlineStatusIdle))
	case "invisible":
		opts = append(opts, gateway.WithOnlineStatus(discord.OnlineStatusInvisible))
	case "offline":
		opts = append(opts, gateway.WithOnlineStatus(discord.OnlineStatusOffline))
	default:
		opts = append(opts, gateway.WithOnlineStatus(discord.OnlineStatusOnline))
	}

	if presence.Activity != "" {
		switch strings.ToLower(presence.ActivityType) {
		case "listening":
			opts = append(opts, gateway.WithListeningActivity(presence.Activity))
		case "watching":
			opts = append(opts, gateway.WithWatchingActivity(presence.Activity))
		case "competing":
			opts = append(opts, gateway.WithCompetingActivity(presence.Activity))
		case "streaming":
			opts = append(opts, gateway.WithStreamingActivity(presence.Activity, ""))
		case "custom":
			opts = append(opts, gateway.WithCustomActivity(presence.Activity))
		case "playing":
			fallthrough
		default:
			opts = append(opts, gateway.WithPlayingActivity(presence.Activity))
		}
	}

	return opts
}
