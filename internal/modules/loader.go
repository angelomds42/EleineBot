package modules

import (
	"fmt"
	"strings"
	"sync"

	"github.com/angelomds42/EleineBot/internal/modules/afk"
	"github.com/angelomds42/EleineBot/internal/modules/android"
	"github.com/angelomds42/EleineBot/internal/modules/lastfm"
	"github.com/angelomds42/EleineBot/internal/modules/medias"
	"github.com/angelomds42/EleineBot/internal/modules/menu"
	"github.com/angelomds42/EleineBot/internal/modules/misc"
	"github.com/angelomds42/EleineBot/internal/modules/moderation"
	"github.com/angelomds42/EleineBot/internal/modules/stickers"
	"github.com/go-telegram/bot"
)

var (
	packageLoadersMutex sync.Mutex
	packageLoaders      = map[string]func(*bot.Bot){
		"afk":        afk.Load,
		"moderation": moderation.Load,
		"lastfm":     lastfm.Load,
		"medias":     medias.Load,
		"menu":       menu.Load,
		"misc":       misc.Load,
		"stickers":   stickers.Load,
		"android":    android.Load,
	}
)

func RegisterHandlers(b *bot.Bot) {
	var wg sync.WaitGroup
	done := make(chan struct{}, len(packageLoaders))
	moduleNames := make([]string, 0, len(packageLoaders))

	for name, loadFunc := range packageLoaders {
		wg.Add(1)

		go func(name string, loadFunc func(*bot.Bot)) {
			defer wg.Done()

			packageLoadersMutex.Lock()
			defer packageLoadersMutex.Unlock()

			loadFunc(b)

			done <- struct{}{}
			moduleNames = append(moduleNames, name)
		}(name, loadFunc)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	for range done {
	}

	joinedModuleNames := strings.Join(moduleNames, ", ")

	fmt.Printf("\033[0;35mModules Loaded:\033[0m %s\n", joinedModuleNames)
}
