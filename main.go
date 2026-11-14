package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"scam-checker-bot/internal/analyzer"
	"scam-checker-bot/internal/bot"
)

func main() {
	ctx := context.Background()

	// ── Configuration from environment variables ────────────────────────────
	telegramToken := mustEnv("TELEGRAM_BOT_TOKEN")
	geminiAPIKey := mustEnv("GEMINI_API_KEY")
	webhookURL := os.Getenv("WEBHOOK_URL") // optional; falls back to polling

	// ── Gemini (Google AI Studio) ───────────────────────────────────────────
	geminiAnalyzer, err := analyzer.NewGeminiAnalyzer(ctx, geminiAPIKey)
	if err != nil {
		log.Fatalf("init gemini analyzer: %v", err)
	}
	defer geminiAnalyzer.Close()
	log.Println("✅ Gemini analyzer ready")

	// ── Web Risk (optional — skipped if ADC lacks permission) ───────────────
	var webRiskChecker *analyzer.WebRiskChecker
	wr, err := analyzer.NewWebRiskChecker(ctx)
	if err != nil {
		log.Printf("⚠️  Web Risk unavailable (continuing without it): %v", err)
	} else {
		webRiskChecker = wr
		defer webRiskChecker.Close()
		log.Println("✅ Web Risk checker ready")
	}

	// ── Telegram ────────────────────────────────────────────────────────────
	tgClient, err := bot.NewTelegramClient(telegramToken)
	if err != nil {
		log.Fatalf("init telegram client: %v", err)
	}
	log.Printf("✅ Telegram bot authorized as @%s", tgClient.BotAPI().Self.UserName)

	handler := bot.NewHandler(tgClient, geminiAnalyzer, webRiskChecker)

	// ── Start: webhook or long-polling ──────────────────────────────────────
	if webhookURL != "" {
		startWebhook(ctx, tgClient, handler, webhookURL)
	} else {
		startPolling(ctx, tgClient, handler)
	}
}

// startPolling uses long-polling (simple local dev setup).
func startPolling(ctx context.Context, tg *bot.TelegramClient, h *bot.Handler) {
	log.Println("🔄 Starting long-polling mode…")
	updates := tg.GetUpdatesChan()
	for update := range updates {
		h.HandleUpdate(ctx, update)
	}
}

// startWebhook registers the webhook and serves incoming updates via HTTP.
func startWebhook(ctx context.Context, tg *bot.TelegramClient, h *bot.Handler, webhookURL string) {
	if err := tg.SetWebhook(webhookURL); err != nil {
		log.Fatalf("set webhook: %v", err)
	}
	log.Printf("🌐 Webhook registered: %s", webhookURL)

	port := envOr("PORT", "8080")
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var update tgbotapi.Update
		if err := json.Unmarshal(body, &update); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		h.HandleUpdate(ctx, update)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	addr := ":" + port
	log.Printf("🚀 HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

// mustEnv returns the value of an environment variable or exits with a message.
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %q is not set", key)
	}
	return v
}

// envOr returns the environment variable value or a default.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
