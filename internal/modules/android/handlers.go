package android

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/angelomds42/EleineBot/internal/database/cache"
	"github.com/angelomds42/EleineBot/internal/localization"
	"github.com/angelomds42/EleineBot/internal/utils"
)

const (
	devicesJSONURL   = "https://raw.githubusercontent.com/androidtrackers/certified-android-devices/master/by_device.json"
	cacheExpiry      = 24 * time.Hour
	redisCacheKey    = "device:cache"
	redisLastUpdated = "device:last_updated"
)

type Device struct {
	Name  string `json:"name"`
	Model string `json:"model"`
	Brand string `json:"brand"`
}

type DeviceResponse map[string][]Device

var (
	client     = &http.Client{Timeout: 10 * time.Second}
	memLock    sync.RWMutex
	memDevices DeviceResponse
	memUpdated time.Time
	nameIndex  = make(map[string][]Device)
	indexOnce  sync.Once
	indexLock  sync.RWMutex
)

func isExpired(t time.Time) bool {
	return time.Since(t) > cacheExpiry
}

func loadCache(ctx context.Context) (DeviceResponse, time.Time) {
	memLock.RLock()
	if memDevices != nil && !isExpired(memUpdated) {
		defer memLock.RUnlock()
		return memDevices, memUpdated
	}
	memLock.RUnlock()

	// Redis fallback
	dataStr, err := cache.GetCache(redisCacheKey)
	if err != nil && err.Error() != "redis: nil" {
		slog.Warn("Redis GET cache failed", "error", err)
		return nil, time.Time{}
	}
	if dataStr == "" {
		return nil, time.Time{}
	}

	tsStr, err := cache.GetCache(redisLastUpdated)
	if err != nil && err.Error() != "redis: nil" {
		slog.Warn("Redis GET timestamp failed", "error", err)
		return nil, time.Time{}
	}
	if tsStr == "" {
		return nil, time.Time{}
	}

	ts, err := time.Parse(time.RFC3339, tsStr)
	if err != nil || isExpired(ts) {
		return nil, time.Time{}
	}

	var devices DeviceResponse
	if err := json.Unmarshal([]byte(dataStr), &devices); err != nil {
		slog.Error("Cache unmarshal failed", "error", err)
		return nil, time.Time{}
	}

	memLock.Lock()
	memDevices = devices
	memUpdated = ts
	memLock.Unlock()

	return devices, ts
}

func persistCache(devices DeviceResponse) {
	data, err := json.Marshal(devices)
	if err != nil {
		slog.Error("JSON marshal failed", "error", err)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)

	if err := cache.SetCache(redisCacheKey, string(data), cacheExpiry); err != nil {
		slog.Warn("Redis SET cache failed", "error", err)
	}
	if err := cache.SetCache(redisLastUpdated, now, cacheExpiry); err != nil {
		slog.Warn("Redis SET timestamp failed", "error", err)
	}

	parsed, _ := time.Parse(time.RFC3339, now)
	memLock.Lock()
	memDevices = devices
	memUpdated = parsed
	memLock.Unlock()

	// reset index build
	indexOnce = sync.Once{}
	indexOnce.Do(func() { buildIndex(devices) })
}

func buildIndex(devices DeviceResponse) {
	indexLock.Lock()
	defer indexLock.Unlock()
	nameIndex = make(map[string][]Device)

	for codename, list := range devices {
		key := strings.ToLower(codename)
		for _, d := range list {
			lowerName := strings.ToLower(d.Name)
			nameIndex[lowerName] = append(nameIndex[lowerName], d)
			for _, w := range strings.Fields(lowerName) {
				nameIndex[w] = append(nameIndex[w], d)
			}
			nameIndex[key] = append(nameIndex[key], d)
		}
	}
}

func fetchDevicesData(ctx context.Context) (DeviceResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, devicesJSONURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	var devs DeviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&devs); err != nil {
		return nil, err
	}
	return devs, nil
}

func searchByName(term string) []Device {
	indexLock.RLock()
	defer indexLock.RUnlock()
	lower := strings.ToLower(term)

	if list, ok := nameIndex[lower]; ok {
		return list
	}
	seen := map[string]Device{}
	for _, w := range strings.Fields(lower) {
		if list, ok := nameIndex[w]; ok {
			for _, d := range list {
				seen[d.Model] = d
			}
		}
	}
	res := make([]Device, 0, len(seen))
	for _, d := range seen {
		res = append(res, d)
	}
	return res
}

func searchDevice(ctx context.Context, term string) ([]Device, DeviceResponse, error) {
	term = strings.ToLower(term)
	devices, _ := loadCache(ctx)
	if devices == nil {
		slog.Info("Refreshing devices cache")
		fresh, err := fetchDevicesData(ctx)
		if err != nil {
			slog.Error("Fetch failed", "error", err)
			return nil, nil, err
		}
		persistCache(fresh)
		devices = fresh
	}
	indexOnce.Do(func() { buildIndex(devices) })
	if list, ok := devices[term]; ok {
		return list, devices, nil
	}
	return searchByName(term), devices, nil
}

func deviceHandler(ctx context.Context, b *bot.Bot, upd *models.Update) {
	i18n := localization.Get(upd)
	if upd.Message == nil {
		return
	}

	parts := strings.Fields(upd.Message.Text)
	if len(parts) < 2 {
		utils.SendMessage(ctx, b, upd.Message.Chat.ID, 0, i18n("device-usage-hint"))
		return
	}

	term := strings.Join(parts[1:], " ")
	matches, all, err := searchDevice(ctx, term)
	if err != nil {
		slog.Error("Search error", "err", err)
		utils.SendMessage(ctx, b, upd.Message.Chat.ID, 0, i18n("device-search-error"))
		return
	}

	if len(matches) == 0 {
		utils.SendMessage(ctx, b, upd.Message.Chat.ID, 0,
			i18n("device-not-found", map[string]interface{}{"searchTerm": term}))
		return
	}

	if len(matches) > 5 {
		matches = matches[:5]
	}

	cmap := make(map[string]string, len(all))
	for codename, list := range all {
		for _, d := range list {
			cmap[d.Model] = codename
		}
	}

	var sb strings.Builder
	sb.WriteString(i18n("device-found") + "\n\n")
	for _, d := range matches {
		cn := cmap[d.Model]
		if cn == "" {
			cn = term
		}
		sb.WriteString(i18n("device-info", map[string]interface{}{
			"device-name":     d.Name,
			"device-model":    d.Model,
			"device-brand":    d.Brand,
			"device-codename": cn,
		}))
		sb.WriteString("\n\n")
	}
	if len(matches) == 5 {
		sb.WriteString(i18n("device-more-results"))
	}

	utils.SendMessage(ctx, b, upd.Message.Chat.ID, upd.Message.ID, sb.String())
}

func Load(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "device", bot.MatchTypeCommand, deviceHandler)

	utils.SaveHelp("android")
}
