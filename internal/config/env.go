package config

import (
	"log"
	"os"

	"github.com/disgoorg/snowflake/v2"
	"github.com/joho/godotenv"
)

type Env struct {
	Token   string
	GuildId snowflake.ID
}

func LoadEnv(environment string) Env {
	switch environment {
	case "production", "prod":
		godotenv.Load(".env")
	case "development", "dev":
		godotenv.Load(".env.development")
	default:
		log.Fatalf("Unknown environment: %s", environment)
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		panic("BOT_TOKEN is required!")
	}

	env := os.Getenv("ENV")
	if env == "" {
		panic("ENV is required!")
	}

	// databaseURL := os.Getenv("DATABASE_URL")
	// if databaseURL == "" {
	// 	panic("DATABASE_URL is required!")
	// }

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return Env{
		Token:   token,
		GuildId: snowflake.GetEnv("GUILD_ID"),
	}
}
