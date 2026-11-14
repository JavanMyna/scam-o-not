package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"scam-checker-bot/internal/analyzer"
	"scam-checker-bot/internal/models"
)

// Handler orchestrates message routing and analysis.
type Handler struct {
	tg      *TelegramClient
	gemini  *analyzer.GeminiAnalyzer
	webrisk *analyzer.WebRiskChecker
}

// NewHandler creates a Handler wiring together the Telegram client, Gemini analyzer,
// and optional Web Risk checker.
func NewHandler(tg *TelegramClient, gem *analyzer.GeminiAnalyzer, wr *analyzer.WebRiskChecker) *Handler {
	return &Handler{tg: tg, gemini: gem, webrisk: wr}
}

// HandleUpdate dispatches a Telegram update to the appropriate handler.
func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	chatID := msg.Chat.ID

	switch {
	case msg.Text == "/start" || msg.Text == "/help":
		h.sendHelp(chatID)

	case msg.Photo != nil:
		h.handlePhoto(ctx, chatID, msg)

	case msg.Text != "":
		h.handleText(ctx, chatID, msg.Text)

	default:
		_ = h.tg.SendText(chatID, "Please send a text message or screenshot to analyze.")
	}
}

// sendHelp sends the welcome / instructions message.
func (h *Handler) sendHelp(chatID int64) {
	help := `👮 *Scam Checker Bot*

Send me:
• A *text message* you received (paste it here)
• A *screenshot* of a suspicious message

I'll analyze it and tell you if it looks like a scam, with specific warning signs and safe steps to verify.`
	if err := h.tg.SendMarkdown(chatID, help); err != nil {
		log.Printf("sendHelp error: %v", err)
	}
}

// handleText analyzes a plain-text message.
func (h *Handler) handleText(ctx context.Context, chatID int64, text string) {
	_ = h.tg.SendText(chatID, "🔍 Analyzing your message…")

	report, err := h.gemini.AnalyzeTextMessage(ctx, text)
	if err != nil {
		log.Printf("gemini text analysis error: %v", err)
		_ = h.tg.SendText(chatID, "❌ Analysis failed. Please try again later.")
		return
	}

	// Enrich with Web Risk if available.
	if h.webrisk != nil {
		threats, err := h.webrisk.CheckURLs(ctx, text)
		if err != nil {
			log.Printf("webrisk check error: %v", err)
		} else {
			report.WarningSigns = analyzer.AppendWebRiskWarnings(report.WarningSigns, threats)
			// Bump score if confirmed threats found.
			for _, t := range threats {
				if t.IsThreat && report.RiskScore < 95 {
					report.RiskScore = 95
					report.RiskLevel = "HIGH"
				}
			}
		}
	}

	if err := h.tg.SendMarkdown(chatID, FormatTelegramReport(report)); err != nil {
		log.Printf("send report error: %v", err)
	}
}

// handlePhoto analyzes a screenshot sent as a Telegram photo.
func (h *Handler) handlePhoto(ctx context.Context, chatID int64, msg *tgbotapi.Message) {
	_ = h.tg.SendText(chatID, "🔍 Analyzing your screenshot…")

	// Use the highest-resolution photo available.
	photos := msg.Photo
	best := photos[len(photos)-1]

	imageBytes, mimeType, err := h.tg.DownloadFile(best.FileID)
	if err != nil {
		log.Printf("download photo error: %v", err)
		_ = h.tg.SendText(chatID, "❌ Could not download the image. Please try again.")
		return
	}

	report, err := h.gemini.AnalyzeImageMessage(ctx, imageBytes, mimeType)
	if err != nil {
		log.Printf("gemini image analysis error: %v", err)
		_ = h.tg.SendText(chatID, "❌ Analysis failed. Please try again later.")
		return
	}

	if err := h.tg.SendMarkdown(chatID, FormatTelegramReport(report)); err != nil {
		log.Printf("send report error: %v", err)
	}
}

// FormatTelegramReport renders an AnalysisReport as a Telegram Markdown string.
func FormatTelegramReport(r *models.AnalysisReport) string {
	var emoji string
	switch r.RiskLevel {
	case "HIGH":
		emoji = "🚨"
	case "SUSPICIOUS":
		emoji = "⚠️"
	default:
		emoji = "✅"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s *Risk Level: %s (%d/100)*\n", emoji, r.RiskLevel, r.RiskScore))
	sb.WriteString(fmt.Sprintf("👤 *Claimed Identity:* %s\n\n", r.ClaimedEntity))

	sb.WriteString("🚩 *WARNING SIGNS:*\n")
	if len(r.WarningSigns) == 0 {
		sb.WriteString("• None detected\n")
	} else {
		for _, sign := range r.WarningSigns {
			sb.WriteString(fmt.Sprintf("• %s\n", sign))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("🛡️ *HOW TO CHECK SAFELY BEFORE ACTING:*\n")
	for i, step := range r.SafeVerificationSteps {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
	}

	return sb.String()
}
