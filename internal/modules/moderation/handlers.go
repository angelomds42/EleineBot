package moderation

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/angelomds42/EleineBot/internal/database"
	"github.com/angelomds42/EleineBot/internal/localization"
	"github.com/angelomds42/EleineBot/internal/utils"
)

func parseUserRestriction(ctx context.Context, b *bot.Bot, msg *models.Message) (userID int64, until int, errMsg string) {
	parts := strings.Fields(msg.Text)

	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil {
		userID = msg.ReplyToMessage.From.ID
		if len(parts) >= 2 {
			dur, err := utils.ParseCustomDuration(parts[1])
			if err == nil {
				until = int(time.Now().Add(dur).Unix())
			}
		}
		return userID, until, ""
	}

	if len(parts) < 2 {
		return 0, 0, "id-required"
	}

	arg := parts[1]

	if id, err := strconv.ParseInt(arg, 10, 64); err == nil {
		userID = id
	} else {
		for _, ent := range msg.Entities {
			if ent.Type == "text_mention" && ent.User != nil {
				off := ent.Offset
				ln := ent.Length
				if off+ln <= len(msg.Text) && msg.Text[off:off+ln] == arg {
					userID = ent.User.ID
					break
				}
			}
		}
	}

	if userID == 0 {
		return 0, 0, "id-invalid"
	}

	if len(parts) >= 3 {
		dur, err := utils.ParseCustomDuration(parts[2])
		if err == nil {
			until = int(time.Now().Add(dur).Unix())
		}
	}

	return userID, until, ""
}

func checkUserAdmin(ctx context.Context, b *bot.Bot, msg *models.Message) bool {
	i18n := localization.Get(&models.Update{Message: msg})
	if !IsAdmin(ctx, b, msg.Chat.ID, msg.From.ID) {
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n("user-not-admin"))
		return false
	}
	return true
}

func checkAdminCallback(ctx context.Context, b *bot.Bot, cb *models.CallbackQuery) bool {
	i18n := localization.Get(&models.Update{CallbackQuery: cb})
	msg := cb.Message.Message
	if !IsAdmin(ctx, b, msg.Chat.ID, cb.From.ID) {
		utils.SendCallbackReply(ctx, b, cb.ID, i18n("user-not-admin"))
		return false
	}
	return true
}

func checkBotAdmin(ctx context.Context, b *bot.Bot, msg *models.Message) bool {
	i18n := localization.Get(&models.Update{Message: msg})
	botID, err := utils.GetBotID(ctx, b)
	if err != nil || (msg.Chat.Type != models.ChatTypePrivate && !IsAdmin(ctx, b, msg.Chat.ID, botID)) {
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n("bot-not-admin", nil))
		return false
	}
	return true
}

func IsAdmin(ctx context.Context, b *bot.Bot, chatID int64, userID int64) bool {
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	return err == nil && (member.Type == models.ChatMemberTypeAdministrator || member.Type == models.ChatMemberTypeOwner)
}

func getUserName(msg *models.Message, userID int64) string {
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil {
		return utils.EscapeHTML(msg.ReplyToMessage.From.FirstName)
	}
	return strconv.FormatInt(userID, 10)
}

func disableableHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	text := i18n("disableables-commands")

	for _, command := range utils.DisableableCommands {
		text += "\n- <code>" + command + "</code>"
	}

	utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, text,
		utils.WithReplyMarkupSend((&models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: i18n("back-button"), CallbackData: "config"},
				},
			},
		})))
}

func disableHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	contains := func(array []string, str string) bool {
		for _, item := range array {
			if item == str {
				return true
			}
		}
		return false
	}

	if !checkUserAdmin(ctx, b, update.Message) {
		return
	}

	if len(strings.Fields(update.Message.Text)) > 1 {
		command := strings.Fields(update.Message.Text)[1]
		if !contains(utils.DisableableCommands, command) {
			utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID,
				i18n("command-not-deactivatable", map[string]interface{}{"command": command}))
			return
		}

		if utils.CheckDisabledCommand(command, update.Message.Chat.ID) {
			utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID,
				i18n("command-already-disabled", map[string]interface{}{"command": command}))
			return
		}

		if err := insertDisabledCommand(update.Message.Chat.ID, command); err != nil {
			slog.Error("Error inserting command", "error", err)
			return
		}

		utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID,
			i18n("command-disabled", map[string]interface{}{"command": command}))
		return
	}

	utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("disable-commands-usage"))
}

func enableHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)

	if !checkUserAdmin(ctx, b, update.Message) {
		return
	}

	if len(strings.Fields(update.Message.Text)) > 1 {
		command := strings.Fields(update.Message.Text)[1]

		if !utils.CheckDisabledCommand(command, update.Message.Chat.ID) {
			utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID,
				i18n("command-already-enabled", map[string]interface{}{"command": command}))
			return
		}

		if err := deleteDisabledCommand(command); err != nil {
			slog.Error("Error deleting command", "error", err)
			return
		}

		utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID,
			i18n("command-enabled", map[string]interface{}{"command": command}))
		return
	}

	utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("enable-commands-usage"))
}

func disabledHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	text := i18n("disabled-commands")

	commands, err := getDisabledCommands(update.Message.Chat.ID)
	if err != nil {
		slog.Error("Error getting disabled commands", "error", err)
		return
	}

	if len(commands) == 0 {
		utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, i18n("no-disabled-commands"))
		return
	}

	for _, command := range commands {
		text += "\n- <code>" + command + "</code>"
	}

	utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, text)
}

func languageMenuCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)

	if !checkAdminCallback(ctx, b, update.CallbackQuery) {
		return
	}

	buttons := make([][]models.InlineKeyboardButton, 0, len(database.AvailableLocales))
	for _, lang := range database.AvailableLocales {
		loaded, ok := localization.LangBundles[lang]
		if !ok {
			slog.Error("Language not found in the cache",
				"lang", lang,
				"availableLocales", database.AvailableLocales)
			os.Exit(1)
		}
		languageFlag, _, _ := loaded.FormatMessage("language-flag")
		languageName, _, _ := loaded.FormatMessage("language-name")

		buttons = append(buttons, []models.InlineKeyboardButton{{
			Text:         languageFlag + languageName,
			CallbackData: fmt.Sprintf("setLang %s", lang),
		}})
	}

	utils.EditMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID,
		update.CallbackQuery.Message.Message.ID,
		i18n("language-menu", map[string]any{
			"languageFlag": i18n("language-flag"),
			"languageName": i18n("language-name"),
		}),
		utils.WithReplyMarkup(&models.InlineKeyboardMarkup{InlineKeyboard: buttons}))
}

func setLanguageCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	lang := strings.ReplaceAll(update.CallbackQuery.Data, "setLang ", "")

	if !checkAdminCallback(ctx, b, update.CallbackQuery) {
		return
	}

	dbQuery := "UPDATE groups SET language = ? WHERE id = ?;"
	if update.CallbackQuery.Message.Message.Chat.Type == models.ChatTypePrivate {
		dbQuery = "UPDATE users SET language = ? WHERE id = ?;"
	}

	if _, err := database.DB.Exec(dbQuery, lang, update.CallbackQuery.Message.Message.Chat.ID); err != nil {
		slog.Error("Couldn't update language",
			"ChatID", update.CallbackQuery.Message.Message.ID,
			"Error", err.Error())
	}

	callbackData := "config"
	if update.CallbackQuery.Message.Message.Chat.Type == models.ChatTypePrivate {
		callbackData = "start"
	}

	utils.EditMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID,
		update.CallbackQuery.Message.Message.ID,
		i18n("language-changed"),
		utils.WithReplyMarkup(&models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: i18n("back-button"), CallbackData: callbackData},
				},
			},
		}))
}

func createConfigKeyboard(i18n func(string, ...map[string]any) string) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         i18n("medias"),
					CallbackData: "mediaConfig",
				},
			},
			{
				{
					Text:         i18n("language-flag") + i18n("language-button"),
					CallbackData: "languageMenu",
				},
			},
		},
	}
}

func configHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	utils.SendMessage(ctx, b, update.Message.Chat.ID, update.Message.ID,
		i18n("config-message"),
		utils.WithReplyMarkupSend(createConfigKeyboard(i18n)),
	)
}

func configCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	utils.EditMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID,
		update.CallbackQuery.Message.Message.ID,
		i18n("config-message"),
		utils.WithReplyMarkup(createConfigKeyboard(i18n)),
	)
}

func getMediaConfig(chatID int64) (bool, bool, error) {
	var mediasCaption, mediasAuto bool
	err := database.DB.QueryRow("SELECT mediasCaption, mediasAuto FROM groups WHERE id = ?;", chatID).Scan(&mediasCaption, &mediasAuto)
	return mediasCaption, mediasAuto, err
}

func mediaConfigCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	mediasCaption, mediasAuto, err := getMediaConfig(update.CallbackQuery.Message.Message.Chat.ID)
	if err != nil {
		slog.Error("Couldn't query media config",
			"ChatID", update.CallbackQuery.Message.Message.Chat.ID,
			"Error", err.Error())
		return
	}

	if !checkAdminCallback(ctx, b, update.CallbackQuery) {
		return
	}

	configType := strings.ReplaceAll(update.CallbackQuery.Data, "mediaConfig ", "")
	if configType != "mediaConfig" {
		query := fmt.Sprintf("UPDATE groups SET %s = ? WHERE id = ?;", configType)
		switch configType {
		case "mediasCaption":
			mediasCaption = !mediasCaption
			_, err = database.DB.Exec(query, mediasCaption, update.CallbackQuery.Message.Message.Chat.ID)
		case "mediasAuto":
			mediasAuto = !mediasAuto
			_, err = database.DB.Exec(query, mediasAuto, update.CallbackQuery.Message.Message.Chat.ID)
		}
		if err != nil {
			slog.Error("Error updating media config", "error", err)
			return
		}
	}

	i18n := localization.Get(update)
	state := func(flag bool) string {
		if flag {
			return "✅"
		}
		return "☑️"
	}

	buttons := [][]models.InlineKeyboardButton{
		{
			{Text: i18n("caption-button"), CallbackData: "ieConfig mediasCaption"},
			{Text: state(mediasCaption), CallbackData: "mediaConfig mediasCaption"},
		},
		{
			{Text: i18n("automatic-button"), CallbackData: "ieConfig mediasAuto"},
			{Text: state(mediasAuto), CallbackData: "mediaConfig mediasAuto"},
		},
		{
			{Text: i18n("back-button"), CallbackData: "config"},
		},
	}

	utils.EditMessage(ctx, b, update.CallbackQuery.Message.Message.Chat.ID,
		update.CallbackQuery.Message.Message.ID,
		i18n("config-medias"),
		utils.WithReplyMarkup(&models.InlineKeyboardMarkup{InlineKeyboard: buttons}))
}

func explainConfigCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	i18n := localization.Get(update)
	ieConfig := strings.ReplaceAll(update.CallbackQuery.Data, "ieConfig medias", "")
	utils.SendCallbackReply(ctx, b, update.CallbackQuery.ID, i18n("ieConfig-"+ieConfig))
}

type restrictionFunc func(ctx context.Context, b *bot.Bot, msg *models.Message, userID int64, until int) error

func newRestrictionHandler(name string, action restrictionFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		msg := update.Message
		i18n := localization.Get(update)

		if !checkUserAdmin(ctx, b, msg) || !checkBotAdmin(ctx, b, msg) {
			return
		}

		userID, until, errMsg := parseUserRestriction(ctx, b, msg)
		if errMsg != "" {
			utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n(name+"-id"))
			return
		}

		if err := action(ctx, b, msg, userID, until); err != nil {
			utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n(name+"-failed"))
			return
		}

		respKey := name + "-success"
		pastAction, ok := map[string]string{"mute": "Muted", "unmute": "Unmuted", "ban": "Banned", "unban": "Unbanned", "delete": "Deleted"}[name]
		if !ok {
			pastAction = cases.Title(language.Und).String(name)
		}
		respData := map[string]interface{}{"user" + pastAction + "FirstName": getUserName(msg, userID)}
		if until > 0 {
			expr := name + "-success-temp"
			respKey = expr
			respData["untilDate"] = time.Unix(int64(until), 0).Format("02/01/2006 15:04")
		}
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n(respKey, respData))
	}
}

func unmuteAction(ctx context.Context, b *bot.Bot, msg *models.Message, userID int64, until int) error {
	chat, err := b.GetChat(ctx, &bot.GetChatParams{ChatID: msg.Chat.ID})
	if err != nil {
		return err
	}
	params := &bot.RestrictChatMemberParams{ChatID: msg.Chat.ID, UserID: userID, Permissions: chat.Permissions}
	_, err = b.RestrictChatMember(ctx, params)
	return err
}

func banAction(ctx context.Context, b *bot.Bot, msg *models.Message, userID int64, until int) error {
	revoke := false
	params := &bot.BanChatMemberParams{ChatID: msg.Chat.ID, UserID: userID, RevokeMessages: revoke}
	if until > 0 {
		params.UntilDate = until
	}
	_, err := b.BanChatMember(ctx, params)
	return err
}

func unbanAction(ctx context.Context, b *bot.Bot, msg *models.Message, userID int64, until int) error {
	_, err := b.UnbanChatMember(ctx, &bot.UnbanChatMemberParams{ChatID: msg.Chat.ID, UserID: userID})
	return err
}

func deleteAction(ctx context.Context, b *bot.Bot, msg *models.Message, userID int64, until int) error {
	if msg.ReplyToMessage == nil {
		return fmt.Errorf("no reply message")
	}
	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{ChatID: msg.Chat.ID, MessageID: msg.ReplyToMessage.ID})
	return err
}

func muteAction(ctx context.Context, b *bot.Bot, msg *models.Message, userID int64, until int) error {
	permissions := &models.ChatPermissions{
		CanSendMessages:       false,
		CanSendAudios:         false,
		CanSendDocuments:      false,
		CanSendPhotos:         false,
		CanSendVideos:         false,
		CanSendVideoNotes:     false,
		CanSendVoiceNotes:     false,
		CanSendPolls:          false,
		CanSendOtherMessages:  false,
		CanAddWebPagePreviews: false,
		CanChangeInfo:         false,
		CanInviteUsers:        false,
		CanPinMessages:        false,
	}
	params := &bot.RestrictChatMemberParams{ChatID: msg.Chat.ID, UserID: userID, Permissions: permissions}
	if until > 0 {
		params.UntilDate = until
	}
	_, err := b.RestrictChatMember(ctx, params)
	return err

}

func Load(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "languageMenu", bot.MatchTypeExact, languageMenuCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "setLang", bot.MatchTypeContains, setLanguageCallback)
	b.RegisterHandler(bot.HandlerTypeMessageText, "config", bot.MatchTypeCommand, configHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "config", bot.MatchTypeExact, configCallback)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "mediaConfig", bot.MatchTypeContains, mediaConfigCallback)
	b.RegisterHandler(bot.HandlerTypeMessageText, "disableable", bot.MatchTypeCommand, disableableHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "disable", bot.MatchTypeCommand, disableHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "enable", bot.MatchTypeCommand, enableHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "disabled", bot.MatchTypeCommand, disabledHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "disableable", bot.MatchTypeCommand, disableableHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "ieConfig", bot.MatchTypeExact, explainConfigCallback)
	b.RegisterHandler(bot.HandlerTypeMessageText, "mute", bot.MatchTypeCommand, newRestrictionHandler("mute", muteAction))
	b.RegisterHandler(bot.HandlerTypeMessageText, "unmute", bot.MatchTypeCommand, newRestrictionHandler("unmute", unmuteAction))
	b.RegisterHandler(bot.HandlerTypeMessageText, "ban", bot.MatchTypeCommand, newRestrictionHandler("ban", banAction))
	b.RegisterHandler(bot.HandlerTypeMessageText, "unban", bot.MatchTypeCommand, newRestrictionHandler("unban", unbanAction))
	b.RegisterHandler(bot.HandlerTypeMessageText, "del", bot.MatchTypeCommand, newRestrictionHandler("delete", deleteAction))

	utils.SaveHelp("moderation")
}
