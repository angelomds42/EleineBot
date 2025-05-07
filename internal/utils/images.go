package utils

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log/slog"
	"os"
	"regexp"

	"github.com/anthonynsimon/bild/transform"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var emojiRegex = regexp.MustCompile(`[\x{1F000}-\x{1FAFF}]|[\x{2600}-\x{27BF}]|\x{200D}|[\x{FE00}-\x{FE0F}]|[\x{E0020}-\x{E007F}]|[\x{1F1E6}-\x{1F1FF}][\x{1F1E6}-\x{1F1FF}]`)

func ResizeSticker(input []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}

	// Convert the image to NRGBA to ensure PNG compatibility
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(bounds)
	draw.Draw(nrgba, bounds, img, bounds.Min, draw.Src)

	resizedImg := transform.Resize(nrgba, 512, 512, transform.Lanczos)

	var buf bytes.Buffer
	if err := png.Encode(&buf, resizedImg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func processThumbnailImage(img image.Image) ([]byte, error) {
	var buf bytes.Buffer

	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()
	if originalWidth > 320 || originalHeight > 320 {
		aspectRatio := float64(originalWidth) / float64(originalHeight)
		var newWidth, newHeight int
		if originalWidth > originalHeight {
			newWidth = 320
			newHeight = int(float64(newWidth) / aspectRatio)
		} else {
			newHeight = 320
			newWidth = int(float64(newHeight) * aspectRatio)
		}
		img = transform.Resize(img, newWidth, newHeight, transform.Linear)
	}

	quality := 100
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}

	for int64(buf.Len()) > 200*1024 && quality > 10 {
		quality -= 10
		buf.Reset()
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func ResizeThumbnail(input []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	return processThumbnailImage(img)
}

func CheckStickerSetCount(ctx context.Context, b *bot.Bot, name string) bool {
	set, err := b.GetStickerSet(ctx, &bot.GetStickerSetParams{Name: name})
	if err != nil {
		return false
	}
	return len(set.Stickers) > 120
}

func GenerateStickerSetName(ctx context.Context, b *bot.Bot, update *models.Update) (shortName, title string) {
	botInfo, err := b.GetMe(ctx)
	if err != nil {
		slog.Error("generateStickerSetName: getMe failed", "error", err)
		os.Exit(1)
	}

	prefix := "a_"
	suffix := fmt.Sprintf("%d_by_%s", update.Message.From.ID, botInfo.Username)

	nameTitle := update.Message.From.FirstName
	if u := update.Message.From.Username; u != "" {
		nameTitle = "@" + u
	}
	if len(nameTitle) > 35 {
		nameTitle = nameTitle[:35]
	}
	title = fmt.Sprintf("%s's Eleine", nameTitle)
	shortName = prefix + suffix

	for i := 0; CheckStickerSetCount(ctx, b, shortName); i++ {
		shortName = fmt.Sprintf("%s%d_%s", prefix, i, suffix)
	}
	return
}

func ExtractEmojis(text string, fallback string) []string {
	emojis := emojiRegex.FindAllString(text, -1)
	if len(emojis) == 0 {
		return []string{fallback}
	}
	return emojis
}
