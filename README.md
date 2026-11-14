# Scam-O-Not 🛡️

A modern, responsive web application that helps customers recognize suspicious support messages and social engineering scams using **Google Gemini AI**.

Users can paste text messages or upload/paste screenshots to detect urgency tactics, brand impersonation, deceptive links, and credential harvesting—complete with a visual risk meter and actionable safe verification steps.

---

## ✨ Features

- **📝 Text Message Analysis**: Paste any suspicious SMS, chat message, or email.
- **🖼️ Multimodal Screenshot Analysis**: Drag-and-drop, browse, or paste screenshots directly from your clipboard (`Ctrl+V`).
- **⚡ Preloaded Test Samples**: One-click test scenarios for Netflix Phishing, PayPal Security Alerts, Amazon Orders, and Legitimate Bank OTPs.
- **🤖 Powered by Google Gemini**: Leverages `gemini-3.5-flash` with strict structured JSON schema enforcement.
- **🎯 Comprehensive Risk Evaluation**:
  - **Risk Score (0–100)**: Visual color-coded circular progress gauge.
  - **Risk Level**: Clear classification (`HIGH RISK 🚨`, `SUSPICIOUS ⚠️`, `LOW RISK ✅`).
  - **Claimed Entity**: Identifies the impersonated brand or company.
  - **Observed Warning Signs**: Specific breakdown of detected manipulation tactics.
  - **Safe Verification Steps**: Concrete, actionable guidance on verifying claims safely without clicking links.
- **📋 One-Click Copy Report**: Copies a cleanly formatted summary for customer support or fraud escalation.
- **🔒 Client-Side Privacy**: Your Gemini API key is stored securely in browser `localStorage` and never transmitted to intermediate servers.

---

## 🚀 Quick Start

### 1. Prerequisites
- Any modern web browser (Chrome, Edge, Firefox, Safari).
- A free Google Gemini API key from [Google AI Studio](https://aistudio.google.com/app/apikey).

### 2. Running Locally

You can run the web app in either of two ways:

#### Option A: Direct File Open
Simply double-click [`index.html`](index.html) in your file explorer to open it in your default browser.

#### Option B: Local Web Server (Recommended)
Using Python (preinstalled on most systems):
```bash
# Start a local static server
python -m http.server 8080

# Or using Node.js / npx:
# npx serve .
```
Then open your browser at:
👉 **`http://localhost:8080`**

---

## 🔑 Configuring Your API Key

1. Click the **🔑 API Key** button in the top-right header.
2. Paste your Google AI Studio API key (`AIzaSy...` or `AQ.Ab8...`).
3. Click **Save Key**. The key is saved locally in your browser's `localStorage`.

---

## 📁 Project Structure

```text
scam-o-not/
├── index.html       # Semantic HTML5 layout (tabs, dropzone, dashboard, modal)
├── style.css        # Modern cybersecurity dark theme, glassmorphism, radar animations
├── app.js           # Multimodal Gemini API integration, JSON schema parser, UI logic
└── README.md        # Project documentation
```

---

## 🛡️ Target JSON Schema

The tool prompts the Gemini model using structured output constraints adhering strictly to:

```json
{
  "risk_score": 95,
  "risk_level": "HIGH",
  "claimed_entity": "Netflix Support",
  "warning_signs": [
    "Threatens immediate cancellation within 24 hours to induce panic.",
    "The domain 'netflix-billing-update.top' is not an official Netflix property.",
    "Generic greeting ('Dear Customer') instead of addressing the account holder by name."
  ],
  "safe_verification_steps": [
    "Never click links inside billing reminder messages.",
    "Open the official Netflix app or navigate directly to netflix.com in your browser.",
    "Go to Account > Billing Details to check your genuine payment status."
  ]
}
```

---

## 📸 Example Dashboard Output

```text
🚨 Risk Level: HIGH (95/100)
👤 Claimed Identity: Netflix Support

🚩 OBSERVED WARNING SIGNS:
• Threatens immediate account cancellation within 24 hours.
• Fake billing domain: netflix-billing-update.top is not an official domain.
• Generic greeting ('Dear Customer') instead of addressing account holder by name.
• Unsolicited request to enter payment information via an external link.

🛡️ HOW TO CHECK SAFELY BEFORE ACTING:
1. Do not click links or buttons inside the reminder message.
2. Manually open your browser and type netflix.com or open the official Netflix app.
3. Check your actual payment and subscription status in Account > Billing Details.
4. Contact Netflix customer service via their verified help center if unsure.
```

---

## 📜 License

MIT License. Built for consumer safety and defensive cybersecurity.
