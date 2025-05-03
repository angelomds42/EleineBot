package utils

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/angelomds42/EleineBot/internal/database"
)

var DisableableCommands []string

func CheckDisabledCommand(command string, chatID int64) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM commandsDisabled WHERE command = ? AND chat_id = ? LIMIT 1);"
	err := database.DB.QueryRow(query, command, chatID).Scan(&exists)
	if err != nil {
		fmt.Printf("Error checking command: %v\n", err)
		return false
	}
	return exists
}

func CheckDisabledMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil || len(strings.Fields(update.Message.Text)) < 1 || update.Message.Chat.Type == models.ChatTypePrivate {
			next(ctx, b, update)
			return
		}

		if len(strings.Fields(update.Message.Text)) < 1 {
			return
		}

		command := strings.Replace(strings.Fields(update.Message.Text)[0], "/", "", 1)
		if CheckDisabledCommand(command, update.Message.Chat.ID) {
			return
		}

		next(ctx, b, update)
	}
}

func ParseCustomDuration(s string) (time.Duration, error) {
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration")
	}
	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, err
	}
	switch unit {
	case 'd':
		return time.Hour * 24 * time.Duration(num), nil
	case 'w':
		return time.Hour * 24 * 7 * time.Duration(num), nil
	case 'm':
		return time.Hour * 24 * 30 * time.Duration(num), nil
	default:
		return 0, fmt.Errorf("invalid unit")
	}
}
