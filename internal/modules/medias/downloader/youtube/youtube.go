package youtube

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot/models"
	"github.com/steino/youtubedl"

	"github.com/angelomds42/EleineBot/internal/config"
	"github.com/angelomds42/EleineBot/internal/modules/medias/downloader"
	"github.com/angelomds42/EleineBot/internal/utils"
)

func getVideoFormat(video *youtubedl.Video, itag int) (*youtubedl.Format, error) {
	formats := video.Formats.Itag(itag)
	if len(formats) == 0 {
		return nil, errors.New("invalid itag")
	}
	return &formats[0], nil
}

func ConfigureYoutubeClient() *youtubedl.Client {
	var ytClient *youtubedl.Client

	if config.Socks5Proxy != "" {
		proxyURL, _ := url.Parse(config.Socks5Proxy)
		client := &http.Client{
			Transport: &http.Transport{
				Proxy:                 http.ProxyURL(proxyURL),
				ResponseHeaderTimeout: 30 * time.Second,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
			},
			Timeout: 120 * time.Second,
		}
		ytClient, err := youtubedl.New(youtubedl.WithHTTPClient(client))
		if err != nil {
			slog.Error("Couldn't create youtube-dl client with proxy",
				"Error", err.Error())
			return nil
		}
		return ytClient
	}

	ytClient, err := youtubedl.New()
	if err != nil {
		slog.Error("Couldn't create youtube-dl client",
			"Error", err.Error())
		return nil
	}

	return ytClient
}

func downloadStream(youtubeClient *youtubedl.Client, video *youtubedl.Video, format *youtubedl.Format) ([]byte, error) {
	for attempt := 1; attempt <= 5; attempt++ {
		stream, _, err := youtubeClient.GetStream(video, format)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		defer stream.Close()

		var buf strings.Builder
		if _, err := io.Copy(&buf, stream); err == nil {
			return []byte(buf.String()), nil
		}

		time.Sleep(3 * time.Second)
	}

	return nil, errors.New("YouTube — Failed after 5 attempts")
}

func Downloader(callbackData []string) ([]byte, *youtubedl.Video, error) {
	youtubeClient := ConfigureYoutubeClient()

	var video *youtubedl.Video
	var err error

	for attempt := 1; attempt <= 5; attempt++ {
		video, err = youtubeClient.GetVideo(callbackData[1], youtubedl.WithClient("ANDROID"))
		if err == nil {
			break
		}
		slog.Warn("GetVideo failed, retrying...",
			"attempt", attempt,
			"error", err.Error())
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return nil, video, err
	}

	itag, _ := strconv.Atoi(callbackData[2])
	format, err := getVideoFormat(video, itag)
	if err != nil {
		return nil, video, err
	}

	fileBytes, err := downloadStream(youtubeClient, video, format)
	if err != nil {
		slog.Error("Could't download video, all attempts failed",
			"Error", err.Error())
		return nil, video, err
	}

	if callbackData[0] == "_vid" {
		audioFormat, err := getVideoFormat(video, 140)
		if err != nil {
			return nil, video, err
		}
		audioBytes, err := downloadStream(youtubeClient, video, audioFormat)
		if err != nil {
			return nil, video, err
		}
		fileBytes, err = downloader.MergeAudioVideoBytes(fileBytes, audioBytes)
		if err != nil {
			slog.Error("Could't merge audio and video",
				"Error", err.Error())
			return nil, video, err
		}
	}

	return fileBytes, video, nil
}

func GetBestQualityVideoStream(formats []youtubedl.Format) *youtubedl.Format {
	var bestFormat youtubedl.Format
	var maxBitrate int

	isDesiredQuality := func(qualityLabel string) bool {
		supportedQualities := []string{"1080p", "720p", "480p", "360p", "240p", "144p"}
		for _, supported := range supportedQualities {
			if strings.Contains(qualityLabel, supported) {
				return true
			}
		}
		return false
	}

	for _, format := range formats {
		if format.Bitrate > maxBitrate && isDesiredQuality(format.QualityLabel) {
			maxBitrate = format.Bitrate
			bestFormat = format
		}
	}

	return &bestFormat
}

func Handle(videoURL string) ([]models.InputMedia, []string) {
	youtubeClient := ConfigureYoutubeClient()
	video, err := youtubeClient.GetVideo(videoURL, youtubedl.WithClient("ANDROID"))
	if err != nil {
		if strings.Contains(err.Error(), "can't bypass age restriction") {
			slog.Warn("Age restricted video",
				"URL", videoURL)
			return nil, []string{}
		}
		slog.Error("Couldn't get video",
			"URL", videoURL,
			"Error", err.Error())
		return nil, []string{}
	}

	videoStream := GetBestQualityVideoStream(video.Formats.Type("video/mp4"))

	format, err := getVideoFormat(video, videoStream.ItagNo)
	if err != nil {
		return nil, []string{}
	}

	fileBytes, err := downloadStream(youtubeClient, video, format)
	if err != nil {
		slog.Error("Couldn't download video stream",
			"Error", err.Error())
		return nil, []string{}
	}

	audioFormat, err := getVideoFormat(video, 140)
	if err != nil {
		slog.Error("Couldn't get audio format",
			"Error", err.Error())
		return nil, []string{}
	}

	audioBytes, err := downloadStream(youtubeClient, video, audioFormat)
	if err != nil {
		slog.Error("Couldn't download audio stream",
			"Error", err.Error())
		return nil, []string{}
	}

	fileBytes, err = downloader.MergeAudioVideoBytes(fileBytes, audioBytes)
	if err != nil {
		slog.Error("Couldn't merge audio and video",
			"Error", err.Error())
		return nil, []string{}
	}

	return []models.InputMedia{&models.InputMediaVideo{
		Media:             "attach://" + utils.SanitizeString(fmt.Sprintf("Eleine-YouTube_%s_%s.mp4", video.Author, video.Title)),
		Width:             video.Formats.Itag(videoStream.ItagNo)[0].Width,
		Height:            video.Formats.Itag(videoStream.ItagNo)[0].Height,
		SupportsStreaming: true,
		MediaAttachment:   bytes.NewBuffer(fileBytes),
	}}, []string{fmt.Sprintf("<b>%s:</b> %s", video.Author, video.Title), video.ID}
}
