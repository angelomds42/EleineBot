package stickers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/angelomds42/EleineBot/internal/localization"
	"github.com/angelomds42/EleineBot/internal/utils"
)

func getFileIDAndType(reply *models.Message) (stickerAction string, stickerType string, fileID string) {
	if document := reply.Document; document != nil {
		fileID = document.FileID
		switch {
		case strings.Contains(document.MimeType, "image"):
			stickerType = "static"
			stickerAction = "resize"
		case strings.Contains(document.MimeType, "tgsticker"):
			stickerType = "animated"
		case strings.Contains(document.MimeType, "video"):
			stickerType = "video"
			stickerAction = "convert"
		}
	} else {
		switch {
		case reply.Photo != nil:
			stickerType = "static"
			stickerAction = "resize"
			fileID = reply.Photo[len(reply.Photo)-1].FileID
		case reply.Video != nil:
			stickerType = "video"
			stickerAction = "convert"
			fileID = reply.Video.FileID
		case reply.Animation != nil:
			stickerType = "video"
			stickerAction = "convert"
			fileID = reply.Animation.FileID
		}
	}

	if replySticker := reply.Sticker; replySticker != nil {
		if replySticker.IsAnimated {
			stickerType = "animated"
		} else if replySticker.IsVideo {
			stickerType = "video"
		} else {
			stickerType = "static"
		}
		fileID = replySticker.FileID
	}

	return stickerAction, stickerType, fileID
}

func getStickerHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	if update.Message.ReplyToMessage == nil {
		utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("get-sticker-no-reply-provided"))
		return
	}

	reply := update.Message.ReplyToMessage.Sticker
	if reply != nil && !reply.IsAnimated {
		file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: reply.FileID})
		if err != nil {
			slog.Error("get file failed", "error", err)
			utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("kang-error"))
			return
		}

		resp, err := http.Get(b.FileDownloadLink(file))
		if err != nil {
			slog.Error("download failed", "error", err)
			utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("kang-error"))
			return
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("read failed", "error", err)
			utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("kang-error"))
			return
		}

		_, err = b.SendDocument(ctx, &bot.SendDocumentParams{
			ChatID: update.Message.Chat.ID,
			Document: &models.InputFileUpload{
				Filename: filepath.Base(b.FileDownloadLink(file)),
				Data:     bytes.NewBuffer(data),
			},
			Caption:                     fmt.Sprintf("<b>Emoji: %s</b>\n<b>ID:</b> <code>%s</code>", reply.Emoji, reply.FileID),
			ParseMode:                   models.ParseModeHTML,
			DisableContentTypeDetection: *bot.True(),
		})
		if err != nil {
			slog.Error("send document failed", "error", err)
			utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("kang-error"))
		}
	}
}

func editStickerError(ctx context.Context, b *bot.Bot, update *models.Update, prog *models.Message, i18n func(string, ...map[string]any) string, msg string, err error) {
	slog.Error(msg, "error", err)
	utils.EditMessage(ctx, b, update.Message.Chat.ID, prog.ID, i18n("kang-error"))
}

func checkStickerSetCount(ctx context.Context, b *bot.Bot, name string) bool {
	set, err := b.GetStickerSet(ctx, &bot.GetStickerSetParams{Name: name})
	if err != nil {
		return false
	}
	return len(set.Stickers) >= 120
}

func sendProgressMessage(ctx context.Context, b *bot.Bot, update *models.Update, i18n func(string, ...map[string]any) string) *models.Message {
	prog, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          update.Message.Chat.ID,
		Text:            i18n("kanging"),
		ParseMode:       models.ParseModeHTML,
		ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
	})
	if err != nil {
		slog.Error("progress message failed", "error", err)
		return nil
	}
	return prog
}

func downloadStickerData(ctx context.Context, b *bot.Bot, update *models.Update, prog *models.Message,
	i18n func(string, ...map[string]any) string, fileID string) ([]byte, *models.File, error) {

	file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		editStickerError(ctx, b, update, prog, i18n, "get file failed", err)
		return nil, nil, err
	}

	resp, err := http.Get(b.FileDownloadLink(file))
	if err != nil {
		editStickerError(ctx, b, update, prog, i18n, "download failed", err)
		return nil, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		editStickerError(ctx, b, update, prog, i18n, "read failed", err)
		return nil, nil, err
	}
	return data, file, nil
}

func processStickerData(ctx context.Context, b *bot.Bot, update *models.Update, prog *models.Message, i18n func(string, ...map[string]any) string, action string, data []byte) ([]byte, error) {
	var processedData []byte
	var err error

	switch action {
	case "resize":
		processedData, err = utils.ResizeSticker(data)
	case "convert":
		utils.EditMessage(ctx, b, update.Message.Chat.ID, prog.ID, i18n("converting-video-to-sticker"))
		processedData, err = convertVideo(data)
	default:
		processedData = data
	}

	if err != nil {
		editStickerError(ctx, b, update, prog, i18n, action+" failed", err)
		return nil, err
	}
	return processedData, nil
}

func extractEmojis(update *models.Update, i18n func(string, ...map[string]any) string) []string {
	re := regexp.MustCompile(`[\x{1F000}-\x{1FAFF}]|[\x{2600}-\x{27BF}]|\x{200D}|[\x{FE00}-\x{FE0F}]|[\x{1F1E6}-\x{1F1FF}]`)
	emoji := re.FindAllString(update.Message.Text, -1)

	if len(emoji) == 0 && update.Message.ReplyToMessage.Sticker != nil {
		emoji = []string{update.Message.ReplyToMessage.Sticker.Emoji}
	}
	if len(emoji) == 0 {
		emoji = []string{"🤔"}
	}
	return emoji
}

func getProcessedFileName(b *bot.Bot, file *models.File, action string) string {
	fileName := filepath.Base(b.FileDownloadLink(file))
	if action == "resize" && !strings.HasSuffix(fileName, ".webp") {
		fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".png"
	}
	return fileName
}

func handleStickerSet(ctx context.Context, b *bot.Bot, update *models.Update, prog *models.Message, i18n func(string, ...map[string]any) string, data []byte, stype, fileName string, emojis []string, shortName, title string) error {
	uploaded, err := b.UploadStickerFile(ctx, &bot.UploadStickerFileParams{
		UserID:        update.Message.From.ID,
		Sticker:       &models.InputFileUpload{Filename: fileName, Data: bytes.NewBuffer(data)},
		StickerFormat: stype,
	})
	if err != nil {
		editStickerError(ctx, b, update, prog, i18n, "upload failed", err)
		return err
	}

	if _, err = b.AddStickerToSet(ctx, &bot.AddStickerToSetParams{
		UserID: update.Message.From.ID,
		Name:   shortName,
		Sticker: models.InputSticker{
			Sticker:   &models.InputFileString{Data: uploaded.FileID},
			Format:    stype,
			EmojiList: emojis,
		},
	}); err != nil {
		utils.EditMessage(ctx, b, update.Message.Chat.ID, prog.ID, i18n("sticker-new-pack"))
		if _, err = b.CreateNewStickerSet(ctx, &bot.CreateNewStickerSetParams{
			UserID:   update.Message.From.ID,
			Name:     shortName,
			Title:    title,
			Stickers: []models.InputSticker{{Sticker: &models.InputFileString{Data: uploaded.FileID}, Format: stype, EmojiList: emojis}},
		}); err != nil {
			editStickerError(ctx, b, update, prog, i18n, "create new set failed", err)
			return err
		}
	}
	return nil
}

func sendSuccessMessage(ctx context.Context, b *bot.Bot, update *models.Update, prog *models.Message, i18n func(string, ...map[string]any) string, shortName string, emojis []string) {
	utils.EditMessage(ctx, b, update.Message.Chat.ID, prog.ID,
		i18n("sticker-stoled", map[string]any{"stickerSetName": shortName, "emoji": strings.Join(emojis, "")}),
		utils.WithReplyMarkup(&models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{{
				{Text: i18n("sticker-view-pack"), URL: fmt.Sprintf("https://t.me/addstickers/%s", shortName)},
			}},
		}),
	)
}

func StickerSetExists(ctx context.Context, b *bot.Bot, name string) bool {
	_, err := b.GetStickerSet(ctx, &bot.GetStickerSetParams{Name: name})
	return err == nil
}

func generateStickerSetName(ctx context.Context, b *bot.Bot, update *models.Update) (shortName, title string) {
	botInfo, err := b.GetMe(ctx)
	if err != nil {
		slog.Error("bot info failed", "error", err)
		return "", ""
	}

	prefix := "a_"
	suffix := fmt.Sprintf("%d_by_%s", update.Message.From.ID, botInfo.Username)

	title = update.Message.From.FirstName
	if u := update.Message.From.Username; u != "" {
		title = "@" + u
	}
	if len(title) > 35 {
		title = title[:35]
	}
	title = fmt.Sprintf("%s's Eleine", title)

	shortName = prefix + suffix
	for i := 0; checkStickerSetCount(ctx, b, shortName); i++ {
		shortName = fmt.Sprintf("%s%d_%s", prefix, i, suffix)
	}
	return
}

func convertVideo(input []byte) ([]byte, error) {
	in, err := os.CreateTemp("", "in_*.mp4")
	if err != nil {
		return nil, err
	}
	defer os.Remove(in.Name())
	if _, err := in.Write(input); err != nil {
		return nil, err
	}

	out := in.Name() + ".webm"
	cmd := exec.Command(
		"ffmpeg", "-loglevel", "quiet",
		"-i", in.Name(),
		"-t", "00:00:03",
		"-vf", "fps=30,scale=512:512:force_original_aspect_ratio=decrease",
		"-c:v", "vp9", "-b:v", "500k", "-y", out,
	)
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	defer os.Remove(out)
	return os.ReadFile(out)
}

func kangStickerHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)

	if update.Message.ReplyToMessage == nil {
		utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("kang-no-reply-provided"))
		return
	}

	prog := sendProgressMessage(ctx, b, update, i18n)
	if prog == nil {
		return
	}

	action, stype, fileID := getFileIDAndType(update.Message.ReplyToMessage)
	if stype == "" {
		utils.EditMessage(ctx, b, update.Message.Chat.ID, prog.ID, i18n("sticker-invalid-media-type"))
		return
	}

	data, file, err := downloadStickerData(ctx, b, update, prog, i18n, fileID)
	if err != nil {
		return
	}

	processedData, err := processStickerData(ctx, b, update, prog, i18n, action, data)
	if err != nil {
		return
	}

	emojis := extractEmojis(update, i18n)
	shortName, title := generateStickerSetName(ctx, b, update)

	utils.EditMessage(ctx, b, update.Message.Chat.ID, prog.ID, i18n("sticker-pack-already-exists"))

	fileName := getProcessedFileName(b, file, action)
	if err := handleStickerSet(ctx, b, update, prog, i18n, processedData, stype, fileName, emojis, shortName, title); err != nil {
		return
	}

	sendSuccessMessage(ctx, b, update, prog, i18n, shortName, emojis)
}

func Load(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "getsticker", bot.MatchTypeCommand, getStickerHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "kang", bot.MatchTypeCommand, kangStickerHandler)

	utils.SaveHelp("stickers")
	utils.DisableableCommands = append(utils.DisableableCommands, "getsticker", "kang")
}
