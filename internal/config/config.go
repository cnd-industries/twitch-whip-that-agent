// Package config loads application settings from environment variables.
package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	// TwitchChannel is the channel whose chat is read. When empty, the
	// Twitch integration is disabled.
	TwitchChannel string `env:"TWITCH_CHANNEL"`

	// Command is the chat command viewers use to whip the agent.
	Command string `env:"WHIP_COMMAND" env-default:"!whip"`
	// RateLimit is the number of whips each viewer may issue per minute.
	RateLimit int `env:"WHIP_RATE_LIMIT" env-default:"3"`
	// Cooldown is the minimum interval between messages sent to the agent.
	// Whips during the cooldown still count and play on the overlay.
	Cooldown time.Duration `env:"WHIP_COOLDOWN" env-default:"15s"`
	// Interrupt makes whips interrupt Claude Code's current turn.
	Interrupt bool `env:"WHIP_INTERRUPT" env-default:"true"`

	// HTTPAddr is the address the overlay server listens on.
	HTTPAddr string `env:"HTTP_ADDR" env-default:"127.0.0.1:8080"`
}

func Load() (Config, error) {
	var config Config
	if err := cleanenv.ReadEnv(&config); err != nil {
		return Config{}, fmt.Errorf("read environment: %w", err)
	}

	if config.RateLimit < 1 {
		return Config{}, fmt.Errorf("WHIP_RATE_LIMIT must be positive, got %d", config.RateLimit)
	}

	return config, nil
}
