package menu

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/angelomds42/EleineBot/internal/localization"
	"github.com/angelomds42/EleineBot/internal/utils"
)

func createStartKeyboard(i18n func(string, ...map[string]any) string) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: i18n("about-button"), CallbackData: "about"},
				{Text: fmt.Sprintf("%s %s", i18n("language-flag"), i18n("language-button")), CallbackData: "languageMenu"},
			},
			{{Text: i18n("help-button"), CallbackData: "helpMenu"}},
		},
	}
}

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	botUser, err := b.GetMe(ctx)
	if err != nil {
		slog.Error("GetMe failed", "error", err)
		return
	}

	chatID := update.Message.Chat.ID
	msgID := update.Message.ID
	fields := strings.Fields(update.Message.Text)
	if len(fields) > 1 && fields[1] == "privacy" {
		privacyHandler(ctx, b, update)
		return
	}

	if update.Message.Chat.Type == models.ChatTypeGroup || update.Message.Chat.Type == models.ChatTypeSupergroup {
		text := i18n("start-message-group", map[string]any{"botName": botUser.FirstName})
		markup := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{{
				{Text: i18n("start-button"), URL: fmt.Sprintf("https://t.me/%s?start=start", botUser.Username)},
			}},
		}
		utils.SendMessage(ctx, b, chatID, msgID, text, utils.WithReplyMarkupSend(markup))
		return
	}

	text := i18n("start-message", map[string]any{"userFirstName": update.Message.From.FirstName, "botName": botUser.FirstName})
	opts := []func(*bot.SendMessageParams){
		utils.WithReplyMarkupSend(createStartKeyboard(i18n)),
	}
	utils.SendMessage(ctx, b, chatID, msgID, text, opts...)
}

func startCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	botUser, err := b.GetMe(ctx)
	if err != nil {
		slog.Error("GetMe failed", "error", err)
		return
	}

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	msgID := update.CallbackQuery.Message.Message.ID
	text := i18n("start-message", map[string]any{
		"userFirstName": update.CallbackQuery.Message.Message.From.FirstName,
		"botName":       botUser.FirstName,
	})
	utils.EditMessage(ctx, b, chatID, msgID, text, utils.WithReplyMarkup(createStartKeyboard(i18n)))
}

func privacyHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	botUser, err := b.GetMe(ctx)
	if err != nil {
		slog.Error("GetMe failed", "error", err)
		return
	}

	chatID := update.Message.Chat.ID
	msgID := update.Message.ID
	if update.Message.Chat.Type == models.ChatTypeGroup || update.Message.Chat.Type == models.ChatTypeSupergroup {
		text := i18n("privacy-policy-group")
		markup := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{{
				{Text: i18n("privacy-policy-button"), URL: fmt.Sprintf("https://t.me/%s?start=privacy", botUser.Username)},
			}},
		}
		utils.SendMessage(ctx, b, chatID, msgID, text, utils.WithReplyMarkupSend(markup))
		return
	}

	text := i18n("privacy-policy-private")
	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{{
			{Text: i18n("about-your-data-button"), CallbackData: "aboutYourData"},
		}},
	}
	utils.SendMessage(ctx, b, chatID, msgID, text, utils.WithReplyMarkupSend(markup))
}

func privacyCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	msgID := update.CallbackQuery.Message.Message.ID
	text := i18n("privacy-policy-private")
	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: i18n("about-your-data-button"), CallbackData: "aboutYourData"}},
			{{Text: i18n("back-button"), CallbackData: "about"}},
		},
	}
	utils.EditMessage(ctx, b, chatID, msgID, text, utils.WithReplyMarkup(markup))
}

func aboutMenuCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	msgID := update.CallbackQuery.Message.Message.ID
	text := i18n("about")
	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: i18n("donation-button"), URL: "https://ko-fi.com/ruizlenato"}},
			{{Text: i18n("privacy-policy-button"), CallbackData: "privacy"}},
			{{Text: i18n("back-button"), CallbackData: "start"}},
		},
	}
	utils.EditMessage(ctx, b, chatID, msgID, text, utils.WithReplyMarkup(markup))
}

func aboutYourDataCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	msgID := update.CallbackQuery.Message.Message.ID
	text := i18n("about-your-data")
	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{{
			{Text: i18n("back-button"), CallbackData: "privacy"},
		}},
	}
	utils.EditMessage(ctx, b, chatID, msgID, text, utils.WithReplyMarkup(markup))
}

func helpMenuCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	msgID := update.CallbackQuery.Message.Message.ID
	text := i18n("help")
	markup := &models.InlineKeyboardMarkup{InlineKeyboard: utils.GetHelpKeyboard(i18n)}
	utils.EditMessage(ctx, b, chatID, msgID, text, utils.WithReplyMarkup(markup))
}

func helpMessageCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	module := strings.TrimPrefix(update.CallbackQuery.Data, "helpMessage ")

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	msgID := update.CallbackQuery.Message.Message.ID
	text := i18n(fmt.Sprintf("%s-help", module))
	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{{
			{Text: i18n("back-button"), CallbackData: "helpMenu"},
		}},
	}
	utils.EditMessage(ctx, b, chatID, msgID, text, utils.WithReplyMarkup(markup))
}

func Load(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommand, startHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "start", bot.MatchTypeExact, startCallback)
	b.RegisterHandler(bot.HandlerTypeMessageText, "privacy", bot.MatchTypeCommand, privacyHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "privacy", bot.MatchTypeExact, privacyCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "about", bot.MatchTypeExact, aboutMenuCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "aboutYourData", bot.MatchTypeExact, aboutYourDataCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "helpMenu", bot.MatchTypeExact, helpMenuCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "helpMessage", bot.MatchTypePrefix, helpMessageCallback)
}
