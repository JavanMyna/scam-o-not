package bot

import (
	"fmt"
	"io"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramClient wraps the Telegram Bot API.
type TelegramClient struct {
	api *tgbotapi.BotAPI
}

// NewTelegramClient creates a TelegramClient using the provided bot token.
func NewTelegramClient(token string) (*TelegramClient, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("telegram bot api: %w", err)
	}
	return &TelegramClient{api: api}, nil
}

// SendMarkdown sends a Markdown-formatted message to a chat.
func (t *TelegramClient) SendMarkdown(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := t.api.Send(msg)
	return err
}

// SendText sends a plain-text message to a chat.
func (t *TelegramClient) SendText(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.api.Send(msg)
	return err
}

// DownloadFile downloads a file from Telegram's servers given its file ID.
// It returns the raw bytes and the MIME type detected from the content.
func (t *TelegramClient) DownloadFile(fileID string) ([]byte, string, error) {
	file, err := t.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, "", fmt.Errorf("get file info: %w", err)
	}

	url := file.Link(t.api.Token)
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return nil, "", fmt.Errorf("download file: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read file body: %w", err)
	}

	mime := http.DetectContentType(data)
	return data, mime, nil
}

// SetWebhook registers the bot's webhook URL with Telegram.
func (t *TelegramClient) SetWebhook(webhookURL string) error {
	wh, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("new webhook: %w", err)
	}
	_, err = t.api.Request(wh)
	return err
}

// GetUpdatesChan returns a channel of incoming updates (long-polling mode).
func (t *TelegramClient) GetUpdatesChan() tgbotapi.UpdatesChannel {
	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 60
	return t.api.GetUpdatesChan(cfg)
}

// BotAPI exposes the underlying BotAPI for webhook handling.
func (t *TelegramClient) BotAPI() *tgbotapi.BotAPI {
	return t.api
}
