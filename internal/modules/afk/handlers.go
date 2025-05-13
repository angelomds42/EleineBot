package afk

import (
	"context"
	"database/sql"
	"log/slog"
	"regexp"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/angelomds42/EleineBot/internal/localization"
	"github.com/angelomds42/EleineBot/internal/utils"
)

func WithLargeMediaPreview(p *bot.SendMessageParams) {
	p.LinkPreviewOptions = &models.LinkPreviewOptions{PreferLargeMedia: bot.True()}
}

func CheckAFKMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		message := getMessageFromUpdate(update)
		if message == nil || message.From == nil {
			next(ctx, b, update)
			return
		}

		if !isGroupChat(message.Chat.Type) || isAFKCommand(message.Text) {
			next(ctx, b, update)
			return
		}

		mentionedUserID := getUserIDFromMessage(message)

		userID := message.From.ID
		userAFK := user_is_away(userID)
		mentionedAFK := user_is_away(mentionedUserID)

		if !userAFK && !mentionedAFK {
			next(ctx, b, update)
			return
		}

		if userAFK {
			handleReturnFromAFK(ctx, b, update, message, userID)
		}

		if mentionedUserID != 0 && mentionedAFK {
			handleMentionedAFK(ctx, b, update, message, mentionedUserID)
		}

		next(ctx, b, update)
	}
}

func getMessageFromUpdate(update *models.Update) *models.Message {
	if update.Message != nil {
		return update.Message
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Message.Type != 1 {
		return update.CallbackQuery.Message.Message
	}
	return nil
}

func isGroupChat(chatType models.ChatType) bool {
	return chatType == models.ChatTypeGroup || chatType == models.ChatTypeSupergroup
}

func isAFKCommand(text string) bool {
	return regexp.MustCompile(`^/\bafk\b|^\bbrb\b`).MatchString(text)
}

func handleReturnFromAFK(ctx context.Context, b *bot.Bot, update *models.Update, message *models.Message, userID int64) {
	i18n := localization.Get(update)
	_, duration, err := get_user_away(userID)
	if err != nil && err != sql.ErrNoRows {
		slog.Error("Couldn't get user away status", "UserID", userID, "Error", err.Error())
		return
	}

	if err := unset_user_away(userID); err != nil {
		slog.Error("Couldn't unset user away status", "UserID", userID, "Error", err.Error())
		return
	}

	humanizedDuration := localization.HumanizeTimeSince(duration, update)

	utils.SendMessage(ctx, b, message.Chat.ID, message.ID,
		i18n("now-available", map[string]interface{}{
			"userID":        userID,
			"userFirstName": utils.EscapeHTML(message.From.FirstName),
			"duration":      humanizedDuration,
		}),
		WithLargeMediaPreview,
	)
}

func handleMentionedAFK(ctx context.Context, b *bot.Bot, update *models.Update, message *models.Message, mentionedUserID int64) {
	i18n := localization.Get(update)
	reason, duration, err := get_user_away(mentionedUserID)
	if err != nil && err != sql.ErrNoRows {
		slog.Error("Couldn't get user away status", "UserID", mentionedUserID, "Error", err.Error())
		return
	}

	humanizedDuration := localization.HumanizeTimeSince(duration, update)

	user, err := b.GetChat(ctx, &bot.GetChatParams{ChatID: mentionedUserID})
	if err != nil {
		slog.Error("Couldn't get user", "UserID", mentionedUserID, "Error", err.Error())
		return
	}

	text := i18n("user-unavailable", map[string]interface{}{
		"userID":        mentionedUserID,
		"userFirstName": utils.EscapeHTML(user.FirstName),
		"duration":      humanizedDuration,
	})

	if reason != "" {
		text += "\n" + i18n("user-unavailable-reason", map[string]interface{}{
			"reason": reason,
		})
	}

	utils.SendMessage(ctx, b, message.Chat.ID, message.ID, text, WithLargeMediaPreview)
}

func getUserIDFromMessage(message *models.Message) int64 {
	if m := message.ReplyToMessage; m != nil && m.From != nil {
		return m.From.ID
	}

	for _, entity := range message.Entities {
		if entity.Type == "text_mention" {
			return entity.User.ID
		}

		if entity.Type == "mention" {
			username := message.Text[entity.Offset : entity.Offset+entity.Length]
			userID, err := getIDFromUsername(username)
			if err == nil {
				return userID
			}

			slog.Error("Couldn't get user ID from username", "Username", username, "Error", err.Error())
		}
	}

	return 0
}

func setAFKHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	reason := extractReason(update.Message.Text)
	err := set_user_away(update.Message.From.ID, reason, time.Now().UTC())
	if err != nil {
		slog.Error("Couldn't set user away status",
			"UserID", update.Message.From.ID,
			"Error", err.Error())
		return
	}

	i18n := localization.Get(update)

	utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID,
		i18n("user-now-unavailable", map[string]interface{}{
			"userFirstName": utils.EscapeHTML(update.Message.From.FirstName),
		}),
	)
}

func extractReason(text string) string {
	matches := regexp.MustCompile(`^(?:brb|\/afk)\s(.+)$`).FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func Load(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "afk", bot.MatchTypeCommand, setAFKHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "brb", bot.MatchTypePrefix, setAFKHandler)

	utils.SaveHelp("afk")
}
