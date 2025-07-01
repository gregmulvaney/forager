package download

import (
	"strconv"

	"github.com/gofiber/fiber/v3/client"
	"go.uber.org/zap"
)

type Downloader struct {
	Client *client.Client
	logger *zap.Logger
}

func (d *Downloader) Download(url string) {
	_, err := d.Client.Get(url, client.Config{})
	if err != nil {
		d.logger.Debug("Failed to get file", zap.Error(err))
	}
}

func (d *Downloader) GetFileSize(url string) (int64, error) {
	resp, err := d.Client.Head(url, client.Config{})
	if err != nil {
		d.logger.Debug("Failed to get file size", zap.Error(err))
		return 0, err
	}

	contentLength := resp.Header("Content-Length")
	if contentLength == "" {
		d.logger.Debug("Content-Length header not found")
		return 0, nil
	}

	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		d.logger.Debug("Failed to parse Content-Length", zap.Error(err))
		return 0, err
	}

	return size, nil
}

func Init(logger *zap.Logger) *Downloader {
	client := client.New()
	return &Downloader{
		Client: client,
		logger: logger,
	}
}
