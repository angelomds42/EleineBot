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

	"github.com/angelomds42/EleineBot/internal/database"
	"github.com/angelomds42/EleineBot/internal/localization"
	"github.com/angelomds42/EleineBot/internal/utils"
)

func parseUser(msg *models.Message, i18n func(string, ...map[string]any) string) (userID int64, errMsg string) {
	parts := strings.Fields(msg.Text)

	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil {
		return msg.ReplyToMessage.From.ID, ""
	}

	if len(parts) < 2 {
		return 0, i18n("id-required")
	}

	arg := parts[1]
	if id, err := strconv.ParseInt(arg, 10, 64); err == nil {
		return id, ""
	}

	for _, entity := range msg.Entities {
		start := entity.Offset
		length := entity.Length
		if start+length > len(msg.Text) {
			continue
		}
		entityText := msg.Text[start : start+length]
		if entityText == arg && entity.Type == "text_mention" && entity.User != nil {
			return entity.User.ID, ""
		}
	}

	return 0, i18n("id-invalid")
}

func parseUserAndDuration(msg *models.Message, i18n func(string, ...map[string]any) string) (userID int64, untilDate int, errMsg string) {
	parts := strings.Fields(msg.Text)

	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil {
		userID = msg.ReplyToMessage.From.ID
		if len(parts) >= 2 {
			durStr := parts[1]
			if dur, err := utils.ParseCustomDuration(durStr); err == nil {
				untilDate = int(time.Now().Add(dur).Unix())
			}
		}
		return userID, untilDate, ""
	}

	if len(parts) < 2 {
		return 0, 0, i18n("id-required")
	}

	arg := parts[1]
	if id, err := strconv.ParseInt(arg, 10, 64); err == nil {
		userID = id
	} else {
		for _, entity := range msg.Entities {
			start := entity.Offset
			length := entity.Length
			if start+length > len(msg.Text) {
				continue
			}
			entityText := msg.Text[start : start+length]
			if entityText == arg && entity.Type == "text_mention" && entity.User != nil {
				userID = entity.User.ID
				break
			}
		}
	}

	if userID == 0 {
		return 0, 0, i18n("id-invalid")
	}

	if len(parts) >= 3 {
		durStr := parts[2]
		if dur, err := utils.ParseCustomDuration(durStr); err == nil {
			untilDate = int(time.Now().Add(dur).Unix())
		}
	}

	return userID, untilDate, ""
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

func banUserHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	i18n := localization.Get(update)

	if !checkUserAdmin(ctx, b, msg) {
		return
	}
	if !checkBotAdmin(ctx, b, msg) {
		return
	}

	userID, untilDate, errMsg := parseUserAndDuration(msg, i18n)
	if errMsg != "" {
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, errMsg)
		return
	}

	revoke := false
	parts := strings.Fields(msg.Text)
	if (msg.ReplyToMessage != nil && len(parts) >= 3 && strings.EqualFold(parts[2], "revoke")) ||
		(msg.ReplyToMessage == nil && len(parts) >= 4 && strings.EqualFold(parts[3], "revoke")) {
		revoke = true
	}

	params := &bot.BanChatMemberParams{
		ChatID:         msg.Chat.ID,
		UserID:         userID,
		RevokeMessages: revoke,
	}
	if untilDate > 0 {
		params.UntilDate = untilDate
	}

	if _, err := b.BanChatMember(ctx, params); err != nil {
		slog.Error("BanChatMember failed", "userID", userID, "error", err)
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n("ban-failed"))
		return
	}

	respKey := "ban-success"
	respData := map[string]interface{}{"userBannedFirstName": getUserName(msg, userID)}
	if untilDate > 0 {
		respKey = "ban-success-temp"
		respData["untilDate"] = time.Unix(int64(untilDate), 0).Format("02/01/2006 15:04")
	}

	utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n(respKey, respData))
}

func unbanUserHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	i18n := localization.Get(update)

	if !checkUserAdmin(ctx, b, msg) {
		return
	}
	if !checkBotAdmin(ctx, b, msg) {
		return
	}

	userID, errMsg := parseUser(msg, i18n)
	if errMsg != "" {
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, errMsg)
		return
	}

	params := &bot.UnbanChatMemberParams{
		ChatID: msg.Chat.ID,
		UserID: userID,
	}
	slog.Info("Unbanning user", "userID", userID, "chatID", msg.Chat.ID)

	if _, err := b.UnbanChatMember(ctx, params); err != nil {
		slog.Error("UnbanChatMember failed", "userID", userID, "error", err)
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n("unban-failed"))
		return
	}

	respKey := "unban-success"
	respData := map[string]interface{}{"userUnbannedFirstName": getUserName(msg, userID)}

	utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n(respKey, respData))
}

func muteUserHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	i18n := localization.Get(update)

	if !checkUserAdmin(ctx, b, msg) {
		return
	}
	if !checkBotAdmin(ctx, b, msg) {
		return
	}

	userID, untilDate, errMsg := parseUserAndDuration(msg, i18n)
	if errMsg != "" {
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, strings.Replace(errMsg, "id", "mute-id", 1))
		return
	}

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

	params := &bot.RestrictChatMemberParams{
		ChatID:      msg.Chat.ID,
		UserID:      userID,
		Permissions: permissions,
	}
	if untilDate > 0 {
		params.UntilDate = untilDate
	}

	if _, err := b.RestrictChatMember(ctx, params); err != nil {
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n("mute-failed"))
		return
	}

	respKey := "mute-success"
	respData := map[string]interface{}{"userMutedFirstName": getUserName(msg, userID)}
	if untilDate > 0 {
		respKey = "mute-success-temp"
		respData["untilDate"] = time.Unix(int64(untilDate), 0).Format("02/01/2006 15:04")
	}

	utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n(respKey, respData))
}

func unmuteUserHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	i18n := localization.Get(update)

	if !checkUserAdmin(ctx, b, msg) {
		return
	}
	if !checkBotAdmin(ctx, b, msg) {
		return
	}

	userID, errMsg := parseUser(msg, i18n)
	if errMsg != "" {
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, strings.Replace(errMsg, "id", "unmute-id", 1))
		return
	}

	chat, err := b.GetChat(ctx, &bot.GetChatParams{ChatID: msg.Chat.ID})
	if err != nil {
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n("unmute-failed"))
		return
	}

	params := &bot.RestrictChatMemberParams{
		ChatID:      msg.Chat.ID,
		UserID:      userID,
		Permissions: chat.Permissions,
	}

	if _, err := b.RestrictChatMember(ctx, params); err != nil {
		utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n("unmute-failed"))
		return
	}

	respKey := "unmute-success"
	respData := map[string]interface{}{"userUnmutedFirstName": getUserName(msg, userID)}

	utils.SendMessage(ctx, b, msg.Chat.ID, msg.ID, i18n(respKey, respData))
}

func deleteMsgHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	i18n := localization.Get(update)
	chatID := msg.Chat.ID

	if !checkUserAdmin(ctx, b, msg) {
		return
	}
	if !checkBotAdmin(ctx, b, msg) {
		return
	}

	if msg.ReplyToMessage == nil {
		utils.SendMessage(ctx, b, chatID, msg.ID, i18n("delete-msg-id-required"))
		return
	}

	params := &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: msg.ReplyToMessage.ID,
	}

	if _, err := b.DeleteMessage(ctx, params); err != nil {
		slog.Error("DeleteMessage failed", "messageID", msg.ReplyToMessage.ID, "error", err)
		utils.SendMessage(ctx, b, chatID, msg.ID, i18n("delete-msg-failed"))
		return
	}

	utils.SendMessage(ctx, b, chatID, msg.ID, i18n("delete-msg-success"))
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
	b.RegisterHandler(bot.HandlerTypeMessageText, "ban", bot.MatchTypeCommand, banUserHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "unban", bot.MatchTypeCommand, unbanUserHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "mute", bot.MatchTypeCommand, muteUserHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "unmute", bot.MatchTypeCommand, unmuteUserHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "del", bot.MatchTypeCommand, deleteMsgHandler)

	utils.SaveHelp("moderation")
}
