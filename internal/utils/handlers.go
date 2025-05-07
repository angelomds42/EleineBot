package utils

import (
	"context"
	"fmt"
	"log/slog"
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

func SendMessage(
	ctx context.Context,
	b *bot.Bot,
	chatID int64,
	replyTo int,
	text string,
	opts ...func(*bot.SendMessageParams),
) {
	params := &bot.SendMessageParams{
		ChatID:             chatID,
		Text:               text,
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: bot.True()},
	}
	if replyTo != 0 {
		params.ReplyParameters = &models.ReplyParameters{MessageID: replyTo}
	}
	for _, opt := range opts {
		opt(params)
	}
	if _, err := b.SendMessage(ctx, params); err != nil {
		slog.Error("utils: SendMessage failed", "chatID", chatID, "error", err)
	}
}

func SendMessageWithResult(
	ctx context.Context,
	b *bot.Bot,
	chatID int64,
	replyTo int,
	text string,
	opts ...func(*bot.SendMessageParams),
) (*models.Message, error) {
	params := &bot.SendMessageParams{
		ChatID:             chatID,
		Text:               text,
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: bot.True()},
	}
	if replyTo != 0 {
		params.ReplyParameters = &models.ReplyParameters{MessageID: replyTo}
	}
	for _, opt := range opts {
		opt(params)
	}
	msg, err := b.SendMessage(ctx, params)
	if err != nil {
		slog.Error("utils: SendMessageWithResult failed", "chatID", chatID, "error", err)
	}
	return msg, err
}

func WithReplyMarkupSend(markup *models.InlineKeyboardMarkup) func(*bot.SendMessageParams) {
	return func(p *bot.SendMessageParams) {
		p.ReplyMarkup = markup
	}
}

func WithReplyTo(messageID int) func(*bot.SendMessageParams) {
	return func(p *bot.SendMessageParams) {
		p.ReplyParameters = &models.ReplyParameters{MessageID: messageID}
	}
}

func EditMessage(
	ctx context.Context,
	b *bot.Bot,
	chatID int64,
	messageID int,
	text string,
	opts ...func(*bot.EditMessageTextParams),
) {
	params := &bot.EditMessageTextParams{
		ChatID:             chatID,
		MessageID:          messageID,
		Text:               text,
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: bot.True()},
	}
	for _, opt := range opts {
		opt(params)
	}
	if _, err := b.EditMessageText(ctx, params); err != nil {
		slog.Error("utils: EditMessage failed", "chatID", chatID, "messageID", messageID, "error", err)
	}
}

func WithReplyMarkup(markup *models.InlineKeyboardMarkup) func(*bot.EditMessageTextParams) {
	return func(p *bot.EditMessageTextParams) {
		p.ReplyMarkup = markup
	}
}

func SendAudio(
	ctx context.Context,
	b *bot.Bot,
	chatID int64,
	replyTo int,
	audio models.InputFile,
	opts ...func(*bot.SendAudioParams),
) (*models.Message, error) {
	params := &bot.SendAudioParams{ChatID: chatID, Audio: audio, ParseMode: models.ParseModeHTML}
	if replyTo != 0 {
		params.ReplyParameters = &models.ReplyParameters{MessageID: replyTo}
	}
	for _, opt := range opts {
		opt(params)
	}
	msg, err := b.SendAudio(ctx, params)
	if err != nil {
		slog.Error("utils: SendAudio failed", "chatID", chatID, "error", err)
	}
	return msg, err
}

func SendVideo(
	ctx context.Context,
	b *bot.Bot,
	chatID int64,
	replyTo int,
	video models.InputFile,
	width, height int,
	opts ...func(*bot.SendVideoParams),
) (*models.Message, error) {
	params := &bot.SendVideoParams{ChatID: chatID, Video: video, Width: width, Height: height, ParseMode: models.ParseModeHTML}
	if replyTo != 0 {
		params.ReplyParameters = &models.ReplyParameters{MessageID: replyTo}
	}
	for _, opt := range opts {
		opt(params)
	}
	msg, err := b.SendVideo(ctx, params)
	if err != nil {
		slog.Error("utils: SendVideo failed", "chatID", chatID, "error", err)
	}
	return msg, err
}

func WithAudioCaption(caption string) func(*bot.SendAudioParams) {
	return func(p *bot.SendAudioParams) {
		p.Caption = caption
	}
}

func WithVideoCaption(caption string) func(*bot.SendVideoParams) {
	return func(p *bot.SendVideoParams) {
		p.Caption = caption
	}
}

func SendCallbackReply(
	ctx context.Context,
	b *bot.Bot,
	callbackID string,
	text string,
) {
	if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callbackID,
		Text:            text,
		ShowAlert:       true,
	}); err != nil {
		slog.Error("utils: SendCallbackReply failed", "callbackID", callbackID, "error", err)
	}
}
