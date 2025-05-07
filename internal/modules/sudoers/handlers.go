package sudoers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/angelomds42/EleineBot/internal/config"
	"github.com/angelomds42/EleineBot/internal/database"
	"github.com/angelomds42/EleineBot/internal/localization"
	"github.com/angelomds42/EleineBot/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var announceMessages = make(map[int64]string)

func announceHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := getMessage(update)
	if msg == nil || !isOwner(update) {
		return
	}

	chatID := msg.Chat.ID
	lang := getLanguage(update)
	if lang == "" {
		announceMessages[chatID] = utils.FormatText(msg.Text, msg.Entities)
		sendLanguageSelector(ctx, b, msg)
		return
	}

	announceToTargets(ctx, b, update, lang)
	delete(announceMessages, chatID)
}

func getMessage(update *models.Update) *models.Message {
	if update.Message != nil {
		return update.Message
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Message
	}
	return nil
}

func isOwner(update *models.Update) bool {
	if update.Message != nil && update.Message.From.ID == config.OwnerID {
		return true
	}
	if update.CallbackQuery != nil && update.CallbackQuery.From.ID == config.OwnerID {
		return true
	}
	return false
}

func getLanguage(update *models.Update) string {
	if update.CallbackQuery != nil {
		return strings.TrimPrefix(update.CallbackQuery.Data, "announce ")
	}
	return ""
}

func sendLanguageSelector(ctx context.Context, b *bot.Bot, msg *models.Message) {
	var buttons [][]models.InlineKeyboardButton
	for _, lang := range database.AvailableLocales {
		bundle, ok := localization.LangBundles[lang]
		if !ok {
			slog.Error("Language bundle not found", "lang", lang)
			continue
		}
		flag, _, _ := bundle.FormatMessage("language-flag")
		name, _, _ := bundle.FormatMessage("language-name")
		buttons = append(buttons, []models.InlineKeyboardButton{{
			Text:         flag + name,
			CallbackData: fmt.Sprintf("announce %s", lang),
		}})
	}

	utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, "Choose a language:",
		utils.WithReplyMarkupSend(&models.InlineKeyboardMarkup{InlineKeyboard: buttons}),
	)
}

func announceToTargets(ctx context.Context, b *bot.Bot, update *models.Update, lang string) {
	msg := getMessage(update)
	if msg == nil {
		return
	}
	chatID := msg.Chat.ID
	text, ok := announceMessages[chatID]
	if !ok {
		return
	}

	fields := strings.Fields(text)
	if len(fields) < 2 {
		return
	}

	announceType := fields[1]
	body := strings.TrimSpace(strings.TrimPrefix(text, fields[0]+" "+announceType))

	query := getQueryForType(announceType, lang)
	rows, err := database.DB.Query(query)
	if err != nil {
		slog.Error("Query announce targets failed", "error", err)
		return
	}
	defer rows.Close()

	var (
		totalCount   int
		successCount int
	)
	for rows.Next() {
		var targetID int64
		totalCount++
		if err := rows.Scan(&targetID); err != nil {
			continue
		}
		utils.SendMessage(ctx, b, targetID, 0, body)
		successCount++
	}

	failedCount := totalCount - successCount

	cb := update.CallbackQuery
	utils.EditMessage(ctx, b,
		cb.Message.Message.Chat.ID,
		cb.Message.Message.ID,
		fmt.Sprintf(
			"<b>Messages sent:</b> <code>%d</code>\n<b>Failed:</b> <code>%d</code>",
			successCount, failedCount,
		),
	)
}

func getQueryForType(announceType, lang string) string {
	switch announceType {
	case "groups":
		return fmt.Sprintf("SELECT id FROM groups WHERE language = '%s';", lang)
	case "users":
		return fmt.Sprintf("SELECT id FROM users WHERE language = '%s';", lang)
	default:
		return fmt.Sprintf(
			"SELECT id FROM users WHERE language = '%s' UNION ALL SELECT id FROM groups WHERE language = '%s';",
			lang, lang,
		)
	}
}

func Load(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "announce", bot.MatchTypeCommand, announceHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "announce", bot.MatchTypePrefix, announceHandler)
}
