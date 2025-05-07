package misc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/url"
	"strconv"
	"strings"

	"github.com/angelomds42/EleineBot/internal/localization"
	"github.com/angelomds42/EleineBot/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const weatherAPIKey = "8de2d8b3a93542c9a2d8b3a935a2c909"

type weatherSearch struct {
	Location struct {
		Latitude  []float64 `json:"latitude"`
		Longitude []float64 `json:"longitude"`
		Address   []string  `json:"address"`
	} `json:"location"`
}

type weatherResult struct {
	ID                      string `json:"id"`
	V3WxObservationsCurrent struct {
		IconCode             int    `json:"iconCode"`
		RelativeHumidity     int    `json:"relativeHumidity"`
		Temperature          int    `json:"temperature"`
		TemperatureFeelsLike int    `json:"temperatureFeelsLike"`
		WindSpeed            int    `json:"windSpeed"`
		WxPhraseLong         string `json:"wxPhraseLong"`
	} `json:"v3-wx-observations-current"`
	V3LocationPoint struct {
		Location struct {
			City   string `json:"city"`
			Locale struct {
				Locale3 any    `json:"locale3"`
				Locale4 string `json:"locale4"`
			} `json:"locale"`
			AdminDistrict  string `json:"adminDistrict"`
			Country        string `json:"country"`
			DisplayContext string `json:"displayContext"`
		} `json:"location"`
	} `json:"v3-location-point"`
}

type googleResp struct {
	Sentences []struct {
		Trans string `json:"trans"`
	} `json:"sentences"`
	Source string `json:"src"`
}

func translateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	chatID := update.Message.Chat.ID
	msgID := update.Message.ID

	text := extractTranslateText(update.Message.Text, update.Message.ReplyToMessage)
	if text == "" {
		utils.SendMessage(ctx, b, chatID, msgID, i18n("translator-no-args-provided"))
		return
	}

	lang := getTranslateLang(text, update.Message.Chat)
	parts := strings.Fields(text)
	if parts[0] == lang {
		text = strings.TrimSpace(strings.TrimPrefix(text, lang))
	}
	if text == "" {
		utils.SendMessage(ctx, b, chatID, msgID, i18n("translator-no-args-provided"))
		return
	}

	src, dst := "auto", lang
	if p := strings.Split(lang, "-"); len(p) > 1 {
		src, dst = p[0], p[1]
	}

	translation, err := fetchGoogleTranslate(src, dst, text)
	if err != nil {
		slog.Error("translate failed", "error", err)
		return
	}

	result := strings.Join(collectTranslations(translation), "")
	unescaped, _ := url.QueryUnescape(result)
	out := fmt.Sprintf("<b>%s</b> -> <b>%s</b>\n<code>%s</code>",
		translation.Source, dst, unescaped)

	utils.SendMessage(ctx, b, chatID, msgID, out)
}

func extractTranslateText(msgText string, reply *models.Message) string {
	if reply != nil {
		base := reply.Text
		if base == "" {
			base = reply.Caption
		}
		if len(msgText) > 4 {
			return msgText[4:] + " " + base
		}
		return base
	}
	if strings.HasPrefix(msgText, "/tr ") {
		return msgText[4:]
	}
	if strings.HasPrefix(msgText, "/translate ") {
		return msgText[11:]
	}
	return ""
}

func getTranslateLang(text string, chat models.Chat) string {
	langs := []string{"af", "sq", "am", "ar", "hy", "as", "ay", "az", "bm", "eu", "be", "bn", "bho", "bs", "bg", "ca", "ceb", "zh", "co", "hr", "cs", "da", "dv", "doi", "nl", "en", "eo", "et", "ee", "fil", "fi", "fr", "fy", "gl", "ka", "de", "el", "gn", "gu", "ht", "ha", "haw", "he", "iw", "hi", "hmn", "hu", "is", "ig", "ilo", "id", "ga", "it", "ja", "jv", "jw", "kn", "kk", "km", "rw", "gom", "ko", "kri", "ku", "ckb", "ky", "lo", "la", "lv", "ln", "lt", "lg", "lb", "mk", "mai", "mg", "ms", "ml", "mt", "mi", "mr", "mni", "lus", "mn", "my", "ne", "no", "ny", "or", "om", "ps", "fa", "pl", "pt", "pa", "qu", "ro", "ru", "sm", "sa", "gd", "nso", "sr", "st", "sn", "sd", "si", "sk", "sl", "so", "es", "su", "sw", "sv", "tl", "tg", "ta", "tt", "te", "th", "ti", "ts", "tr", "tk", "ak", "uk", "ur", "ug", "uz", "vi", "cy", "xh", "yi", "yo", "zu"}
	check := func(s string) bool {
		for _, v := range langs {
			if v == s {
				return true
			}
		}
		return false
	}

	chatLang, err := localization.GetChatLanguage(chat)
	if err != nil {
		chatLang = "en"
	}

	parts := strings.Fields(text)
	if len(parts) > 0 {
		lang := parts[0]
		p := strings.Split(lang, "-")
		if !check(p[0]) {
			lang = strings.Split(chatLang, "-")[0]
		}
		if len(p) > 1 && !check(p[1]) {
			lang = strings.Split(chatLang, "-")[0]
		}
		return lang
	}
	return "en"
}

func fetchGoogleTranslate(src, dst, text string) (*googleResp, error) {
	host := fmt.Sprintf(
		"https://translate.google.com/translate_a/single?client=at&dt=t&dj=1&sl=%s&tl=%s&q=%s",
		src, dst, url.QueryEscape(text),
	)
	devices := []string{"Linux; U; Android 10; Pixel 4", "Linux; U; Android 11; Pixel 5", "Linux; U; Android 12; Pixel 6"}

	resp, err := utils.Request(host, utils.RequestParams{
		Method:  "POST",
		Headers: map[string]string{"User-Agent": fmt.Sprintf("GoogleTranslate/6.28.0 (%s)", devices[rand.Intn(len(devices))])},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gr googleResp
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return nil, err
	}
	return &gr, nil
}

func collectTranslations(gr *googleResp) []string {
	out := make([]string, len(gr.Sentences))
	for i, s := range gr.Sentences {
		out[i] = s.Trans
	}
	return out
}

func weatherHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	chatID, msgID := update.Message.Chat.ID, update.Message.ID

	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		utils.SendMessage(ctx, b, chatID, msgID, i18n("weather-no-location-provided"))
		return
	}
	query := parts[1]

	langTag, err := localization.GetChatLanguage(update.Message.Chat)
	if err != nil {
		langTag = "en-US"
	}
	langCode := formatLang(langTag)

	resp, err := utils.Request("https://api.weather.com/v3/location/search", utils.RequestParams{
		Method: "GET",
		Query: map[string]string{
			"apiKey":   weatherAPIKey,
			"query":    query,
			"language": langCode,
			"format":   "json",
		},
	})
	if err != nil {
		slog.Error("weather search failed", "error", err)
		return
	}
	defer resp.Body.Close()

	var ws weatherSearch
	if err := json.NewDecoder(resp.Body).Decode(&ws); err != nil {
		slog.Error("weather decode failed", "error", err)
		return
	}

	var buttons [][]models.InlineKeyboardButton
	for i := range ws.Location.Address {
		if i >= 5 {
			break
		}
		buttons = append(buttons, []models.InlineKeyboardButton{{
			Text:         ws.Location.Address[i],
			CallbackData: fmt.Sprintf("_weather|%f|%f", ws.Location.Latitude[i], ws.Location.Longitude[i]),
		}})
	}

	utils.SendMessage(ctx, b, chatID, msgID, i18n("weather-select-location"),
		utils.WithReplyMarkupSend(&models.InlineKeyboardMarkup{InlineKeyboard: buttons}),
	)
}

func callbackWeather(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	data := strings.Split(update.CallbackQuery.Data, "|")
	lat, _ := strconv.ParseFloat(data[1], 64)
	lon, _ := strconv.ParseFloat(data[2], 64)

	langTag, err := localization.GetChatLanguage(update.CallbackQuery.Message.Message.Chat)
	if err != nil {
		langTag = "en-US"
	}
	langCode := formatLang(langTag)

	resp, err := utils.Request(
		"https://api.weather.com/v3/aggcommon/v3-wx-observations-current;v3-location-point",
		utils.RequestParams{
			Method: "GET",
			Query: map[string]string{
				"apiKey":   weatherAPIKey,
				"geocode":  fmt.Sprintf("%.3f,%.3f", lat, lon),
				"language": langCode,
				"units":    i18n("measurement-unit"),
				"format":   "json",
			},
		},
	)
	if err != nil {
		slog.Error("weather details failed", "error", err)
		return
	}
	defer resp.Body.Close()

	var wr weatherResult
	if err := json.NewDecoder(resp.Body).Decode(&wr); err != nil {
		slog.Error("weather details decode failed", "error", err)
		return
	}

	parts := []string{wr.V3LocationPoint.Location.Locale.Locale4}
	if v, ok := wr.V3LocationPoint.Location.Locale.Locale3.(string); ok && v != "" {
		parts = append(parts, v)
	}
	parts = append(parts,
		wr.V3LocationPoint.Location.City,
		wr.V3LocationPoint.Location.AdminDistrict,
		wr.V3LocationPoint.Location.Country,
	)
	localName := strings.Join(parts, ", ")

	utils.EditMessage(ctx, b,
		update.CallbackQuery.Message.Message.Chat.ID,
		update.CallbackQuery.Message.Message.ID,
		i18n("weather-details", map[string]any{
			"localname":            localName,
			"temperature":          wr.V3WxObservationsCurrent.Temperature,
			"temperatureFeelsLike": wr.V3WxObservationsCurrent.TemperatureFeelsLike,
			"relativeHumidity":     wr.V3WxObservationsCurrent.RelativeHumidity,
			"windSpeed":            wr.V3WxObservationsCurrent.WindSpeed,
		}),
	)
}

func formatLang(tag string) string {
	parts := strings.Split(tag, "-")
	return parts[0] + "-" + strings.ToUpper(parts[1])
}

func Load(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "weather", bot.MatchTypeCommand, weatherHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "clima", bot.MatchTypeCommand, weatherHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "_weather", bot.MatchTypePrefix, callbackWeather)
	b.RegisterHandler(bot.HandlerTypeMessageText, "translate", bot.MatchTypeCommand, translateHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "tr", bot.MatchTypeCommand, translateHandler)

	utils.SaveHelp("misc")
	utils.DisableableCommands = append(utils.DisableableCommands,
		"tr", "translate", "weather", "clima")
}
