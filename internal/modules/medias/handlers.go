package medias

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/steino/youtubedl"

	"github.com/angelomds42/EleineBot/internal/config"
	"github.com/angelomds42/EleineBot/internal/database"
	"github.com/angelomds42/EleineBot/internal/localization"
	"github.com/angelomds42/EleineBot/internal/modules/medias/downloader"
	"github.com/angelomds42/EleineBot/internal/modules/medias/downloader/bluesky"
	"github.com/angelomds42/EleineBot/internal/modules/medias/downloader/instagram"
	"github.com/angelomds42/EleineBot/internal/modules/medias/downloader/reddit"
	"github.com/angelomds42/EleineBot/internal/modules/medias/downloader/threads"
	"github.com/angelomds42/EleineBot/internal/modules/medias/downloader/tiktok"
	"github.com/angelomds42/EleineBot/internal/modules/medias/downloader/twitter"
	"github.com/angelomds42/EleineBot/internal/modules/medias/downloader/xiaohongshu"
	"github.com/angelomds42/EleineBot/internal/modules/medias/downloader/youtube"
	"github.com/angelomds42/EleineBot/internal/utils"
)

var (
	mediaRegex = regexp.MustCompile(regexMedia)
	cmdRegex   = regexp.MustCompile(`^/(?:s)?dl`)
	handlerMap = []struct {
		pattern *regexp.Regexp
		h       func(string) ([]models.InputMedia, []string)
	}{
		{regexp.MustCompile(`bsky\.app/`), bluesky.Handle},
		{regexp.MustCompile(`instagram\.com/`), instagram.Handle},
		{regexp.MustCompile(`reddit\.com/`), reddit.Handle},
		{regexp.MustCompile(`threads\.net/`), threads.Handle},
		{regexp.MustCompile(`tiktok\.com/`), tiktok.Handle},
		{regexp.MustCompile(`(twitter|x)\.com/`), twitter.Handle},
		{regexp.MustCompile(`(xiaohongshu|xhslink)\.com/`), xiaohongshu.Handle},
	}
	forceRe = regexp.MustCompile(`(tiktok\.com|reddit\.com)`)
)

const (
	regexMedia     = `(?:http(?:s)?://)?(?:m|vm|vt|www|mobile)?(?:.)?(?:(?:instagram|twitter|x|tiktok|reddit|bsky|threads|xiaohongshu|xhslink)\.(?:com|net|app)|youtube\.com/shorts)/(?:\S*)`
	maxSizeCaption = 1024
)

func mediaDownloadHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	text := update.Message.Text
	if !shouldProcessMedia(update, text) {
		return
	}

	link := mediaRegex.FindString(text)
	i18n := localization.Get(update)
	if link == "" {
		utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("no-link-provided"))
		return
	}

	mediaItems, result := fetchMedia(link, text)
	if len(mediaItems) == 0 || mediaItems[0] == nil {
		return
	}

	if skipPreview(mediaItems[0], update.Message.LinkPreviewOptions) {
		return
	}

	mediaItems = limitItems(mediaItems, 10)
	caption := buildCaption(result, link, update.Message.Chat.ID)
	applyCaption(mediaItems[0], caption)

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: update.Message.Chat.ID,
		Action: models.ChatActionUploadDocument,
	})
	replied, err := b.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
		ChatID:          update.Message.Chat.ID,
		Media:           mediaItems,
		ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
	})
	if err != nil {
		return
	}
	_ = downloader.SetMediaCache(replied, result)
}

func shouldProcessMedia(update *models.Update, text string) bool {
	if cmdRegex.MatchString(text) || update.Message.Chat.Type == models.ChatTypePrivate {
		return true
	}
	var mediasAuto bool
	err := database.DB.QueryRow(
		"SELECT mediasAuto FROM groups WHERE id = ?;", update.Message.Chat.ID,
	).Scan(&mediasAuto)
	return err == nil && mediasAuto
}

func fetchMedia(link, text string) ([]models.InputMedia, []string) {
	for _, h := range handlerMap {
		if h.pattern.MatchString(text) {
			force := forceRe.MatchString(text)
			items, result := h.h(link)
			if force || len(result) >= 2 {
				return items, result
			}
		}
	}
	return nil, nil
}

func skipPreview(item models.InputMedia, opts *models.LinkPreviewOptions) bool {
	if opts == nil || (opts.IsDisabled != nil && *opts.IsDisabled) {
		return false
	}
	raw, _ := item.MarshalInputMedia()
	var info struct{ Type, Media string }
	_ = json.Unmarshal(raw, &info)
	return info.Type == "photo"
}

func limitItems(items []models.InputMedia, limit int) []models.InputMedia {
	if len(items) > limit {
		return items[:limit]
	}
	return items
}

func buildCaption(result []string, link string, chatID int64) string {
	caption := ""
	if len(result) >= 2 {
		caption = result[0]
	}
	if utf8.RuneCountInString(caption) > maxSizeCaption {
		caption = downloader.TruncateUTF8Caption(caption, link)
	}
	var showCaption bool
	if err := database.DB.QueryRow(
		"SELECT mediasCaption FROM groups WHERE id = ?;", chatID,
	).Scan(&showCaption); err == nil && !showCaption {
		caption = ""
	}
	if caption == "" {
		caption = fmt.Sprintf("<a href='%s'>🔗 Link</a>", link)
	}
	return caption
}

func applyCaption(item models.InputMedia, caption string) {
	raw, _ := item.MarshalInputMedia()
	var info struct{ Type, Media string }
	_ = json.Unmarshal(raw, &info)
	switch info.Type {
	case "photo":
		m := item.(*models.InputMediaPhoto)
		m.Caption = caption
		m.ParseMode = models.ParseModeHTML
	case "video":
		m := item.(*models.InputMediaVideo)
		m.Caption = caption
		m.ParseMode = models.ParseModeHTML
	}
}

func youtubeDownloadHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	var videoURL string

	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.Text != "" {
		videoURL = update.Message.ReplyToMessage.Text
	} else if parts := strings.Fields(update.Message.Text); len(parts) > 1 {
		videoURL = parts[1]
	} else {
		utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("youtube-no-url"))
		return
	}

	client := youtube.ConfigureYoutubeClient()
	var video *youtubedl.Video
	var err error
	for attempt := 1; attempt <= 10; attempt++ {
		video, err = client.GetVideo(videoURL, youtubedl.WithClient("ANDROID"))
		if err == nil {
			break
		}
		slog.Warn("GetVideo failed, retrying...", "attempt", attempt, "error", err)
		time.Sleep(5 * time.Second)
	}
	if err != nil || video == nil || video.Formats == nil {
		utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("youtube-invalid-url"))
		return
	}

	videoFmt := youtube.GetBestQualityVideoStream(video.Formats.Type("video/mp4"))
	var audioFmt youtubedl.Format
	if fmts := video.Formats.Itag(140); len(fmts) > 0 {
		audioFmt = fmts[0]
	} else {
		audioFmt = video.Formats.WithAudioChannels().Type("audio/mp4")[0]
	}

	infoText := i18n("youtube-video-info", map[string]interface{}{
		"title":     video.Title,
		"author":    video.Author,
		"audioSize": fmt.Sprintf("%.2f", float64(audioFmt.ContentLength)/(1024*1024)),
		"videoSize": fmt.Sprintf("%.2f", float64(videoFmt.ContentLength+audioFmt.ContentLength)/(1024*1024)),
		"duration":  video.Duration.String(),
	})
	keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{{
		{Text: i18n("youtube-download-audio-button"), CallbackData: fmt.Sprintf("_aud|%s|%d|%d|%d|%d",
			video.ID, audioFmt.ItagNo, audioFmt.ContentLength, update.Message.ID, update.Message.From.ID,
		)},
		{Text: i18n("youtube-download-video-button"), CallbackData: fmt.Sprintf("_vid|%s|%d|%d|%d|%d",
			video.ID, videoFmt.ItagNo, videoFmt.ContentLength+audioFmt.ContentLength, update.Message.ID, update.Message.From.ID,
		)},
	}}}

	utils.SendMessage(ctx, b,
		update.Message.Chat.ID, update.Message.ID,
		infoText,
		utils.WithReplyMarkupSend(keyboard),
	)
}

func youtubeDownloadCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	msg := update.CallbackQuery.Message.Message

	parts := strings.Split(update.CallbackQuery.Data, "|")
	requesterID, _ := strconv.ParseInt(parts[5], 10, 64)
	if update.CallbackQuery.From.ID != requesterID {
		utils.SendCallbackReply(ctx, b, update.CallbackQuery.ID, i18n("denied-button-alert"))
		return
	}

	size, _ := strconv.ParseInt(parts[3], 10, 64)
	limit := int64(1_572_864_000)
	if config.BotAPIURL == "" {
		limit = 50 * 1024 * 1024
	}
	if size > limit {
		utils.SendCallbackReply(ctx, b, update.CallbackQuery.ID,
			i18n("video-exceeds-limit", map[string]any{"size": limit}),
		)
		return
	}

	utils.EditMessage(ctx, b,
		msg.Chat.ID, msg.ID,
		i18n("downloading"),
	)

	if sent, err := trySendCachedYoutubeMedia(ctx, b, update, int(partsToInt(parts[4])), parts); sent || err == nil {
		b.DeleteMessage(ctx, &bot.DeleteMessageParams{ChatID: msg.Chat.ID, MessageID: msg.ID})
		return
	}

	fileBytes, video, err := youtube.Downloader(parts)
	if err != nil {
		utils.EditMessage(ctx, b, msg.Chat.ID, msg.ID, i18n("youtube-error"))
		return
	}

	action := models.ChatActionUploadVideo
	if parts[0] == "_aud" {
		action = models.ChatActionUploadVoice
	}
	utils.EditMessage(ctx, b, msg.Chat.ID, msg.ID, i18n("uploading"))
	b.SendChatAction(ctx, &bot.SendChatActionParams{ChatID: msg.Chat.ID, Action: action})

	thumbURL := strings.Replace(video.Thumbnails[len(video.Thumbnails)-1].URL, "sddefault", "maxresdefault", 1)
	thumbBytes, _ := downloader.FetchBytesFromURL(thumbURL)

	caption := fmt.Sprintf("<b>%s:</b> %s", video.Author, video.Title)
	filenameBase := utils.SanitizeString(fmt.Sprintf("Eleine-%s_%s", video.Author, video.Title))

	var replied *models.Message
	if parts[0] == "_aud" {
		replied, err = b.SendAudio(ctx, &bot.SendAudioParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Audio: &models.InputFileUpload{
				Filename: filenameBase,
				Data:     bytes.NewBuffer(fileBytes),
			},
			Caption:         caption,
			Title:           video.Title,
			Performer:       video.Author,
			Thumbnail:       &models.InputFileUpload{Filename: filenameBase, Data: bytes.NewBuffer(thumbBytes)},
			ParseMode:       models.ParseModeHTML,
			ReplyParameters: &models.ReplyParameters{MessageID: partsToInt(parts[4])},
		})
	} else {
		format := video.Formats.Itag(partsToInt(parts[2]))[0]
		replied, err = b.SendVideo(ctx, &bot.SendVideoParams{
			ChatID:            update.CallbackQuery.Message.Message.Chat.ID,
			Video:             &models.InputFileUpload{Filename: filenameBase, Data: bytes.NewBuffer(fileBytes)},
			Width:             format.Width,
			Height:            format.Height,
			Thumbnail:         &models.InputFileUpload{Filename: filenameBase, Data: bytes.NewBuffer(thumbBytes)},
			Caption:           caption,
			ParseMode:         models.ParseModeHTML,
			SupportsStreaming: true,
			ReplyParameters:   &models.ReplyParameters{MessageID: partsToInt(parts[4])},
		})
	}
	if err != nil {
		slog.Error("send media failed", "error", err)
		utils.EditMessage(ctx, b, msg.Chat.ID, msg.ID, i18n("youtube-error"))
		return
	}

	b.DeleteMessage(ctx, &bot.DeleteMessageParams{ChatID: msg.Chat.ID, MessageID: msg.ID})
	if err := downloader.SetYoutubeCache(replied, parts[1]); err != nil {
		slog.Error("cache set failed", "error", err)
	}
}

func trySendCachedYoutubeMedia(ctx context.Context, b *bot.Bot, update *models.Update, messageID int, parts []string) (bool, error) {
	var fileID, caption string
	var err error
	if parts[0] == "_aud" {
		fileID, caption, err = downloader.GetYoutubeCache(parts[1], "audio")
	} else {
		fileID, caption, err = downloader.GetYoutubeCache(parts[1], "video")
	}
	if err == nil {
		if parts[0] == "_aud" {
			utils.SendAudio(ctx, b,
				update.CallbackQuery.Message.Message.Chat.ID, messageID,
				&models.InputFileString{Data: fileID},
				utils.WithAudioCaption(caption),
			)
		} else {
			utils.SendVideo(ctx, b,
				update.CallbackQuery.Message.Message.Chat.ID, messageID,
				&models.InputFileString{Data: fileID},
				0, 0,
			)
		}
		return true, nil
	}
	return false, err
}

func partsToInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func Load(b *bot.Bot) {
	b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, mediaRegex, mediaDownloadHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "ytdl", bot.MatchTypeCommand, youtubeDownloadHandler)
	b.RegisterHandlerRegexp(bot.HandlerTypeCallbackQueryData, regexp.MustCompile(`^(_(vid|aud))`), youtubeDownloadCallback)

	utils.SaveHelp("medias")
}
