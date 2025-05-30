package lastfm

import (
	"github.com/angelomds42/EleineBot/internal/database"
)

/*
	 func lastFMDisabled(update telego.Update) bool {
		var lastFMCommands bool = true
		message := update.Message
		if message.Chat.Type == telego.ChatTypePrivate {
			return lastFMCommands
		}

		database.DB.QueryRow("SELECT lastFMCommands FROM groups WHERE id = ?;", message.Chat.ID).Scan(&lastFMCommands)
		return lastFMCommands
	}
*/

func setLastFMUsername(userID int64, lastFMUsername string) error {
	_, err := database.DB.Exec("UPDATE users SET lastfm_username = ? WHERE id = ?;", lastFMUsername, userID)
	if err != nil {
		return err
	}
	return nil
}

func getUserLastFMUsername(userID int64) (string, error) {
	var lastFMUsername string
	err := database.DB.QueryRow("SELECT lastfm_username FROM users WHERE id = ?;", userID).Scan(&lastFMUsername)
	return lastFMUsername, err
}
