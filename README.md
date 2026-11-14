# Scam Checker Bot 🛡️

A Telegram bot that helps customers recognize suspicious support messages using
**Gemini 1.5 Flash** (Vertex AI) and **Google Web Risk**.

## Features

| Feature | Detail |
|---|---|
| Text analysis | Paste any suspicious message |
| Screenshot analysis | Send a photo — the bot uses multimodal OCR + analysis |
| Web Risk enrichment | Known-malicious URLs are automatically flagged |
| Structured output | Risk score, risk level, warning signs, safe verification steps |

## Quick Start

### Prerequisites

- Go 1.22+
- A Telegram bot token ([BotFather](https://t.me/BotFather))
- A Google Cloud project with **Vertex AI API** and **Web Risk API** enabled
- [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials) configured

### Local Development (long-polling)

```bash
cp .env.example .env
# Edit .env with your credentials
export $(grep -v '^#' .env | xargs)
go run main.go
```

### Docker

```bash
docker build -t scam-checker-bot .
docker run --rm \
  -e TELEGRAM_BOT_TOKEN=... \
  -e GCP_PROJECT_ID=... \
  -e GOOGLE_APPLICATION_CREDENTIALS=/creds/sa.json \
  -v /path/to/sa.json:/creds/sa.json \
  scam-checker-bot
```

### Cloud Run (webhook mode)

```bash
gcloud run deploy scam-checker-bot \
  --image gcr.io/YOUR_PROJECT/scam-checker-bot \
  --set-env-vars TELEGRAM_BOT_TOKEN=...,GCP_PROJECT_ID=...,WEBHOOK_URL=https://YOUR_SERVICE_URL/webhook \
  --allow-unauthenticated
```

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `TELEGRAM_BOT_TOKEN` | ✅ | — | Bot token from BotFather |
| `GCP_PROJECT_ID` | ✅ | — | Google Cloud project ID |
| `GCP_LOCATION` | ❌ | `us-central1` | Vertex AI region |
| `WEBHOOK_URL` | ❌ | — | Omit to use long-polling |
| `PORT` | ❌ | `8080` | HTTP port (webhook mode only) |

## Project Structure

```
scam-checker-bot/
├── Dockerfile
├── main.go                        # Entry point; wires all components
├── go.mod / go.sum
├── internal/
│   ├── models/
│   │   └── report.go              # AnalysisReport struct
│   ├── analyzer/
│   │   ├── gemini.go              # Vertex AI / Gemini 1.5 Flash client
│   │   └── webrisk.go             # Google Web Risk URL checker
│   └── bot/
│       ├── handler.go             # Update routing + report formatting
│       └── telegram.go            # Telegram API client
```

## Example Bot Output

```
🚨 Risk Level: HIGH (90/100)
👤 Claimed Identity: Netflix Support

🚩 WARNING SIGNS:
• Threatens immediate cancellation if payment not updated.
• Domain 'netflix-billing-update.top' is not netflix.com.
• Generic greeting ("Dear Customer") used.
• URL flagged by Google Web Risk as [SOCIAL_ENGINEERING]: http://netflix-billing-update.top/pay

🛡️ HOW TO CHECK SAFELY BEFORE ACTING:
1. Never click links inside billing reminder messages.
2. Open the Netflix app or type netflix.com directly in your browser.
3. Go to Account > Billing Details to verify your actual payment status.
```

## License

MIT
