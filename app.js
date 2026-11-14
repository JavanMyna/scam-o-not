// Default API key placeholder (configure via the UI settings modal)
const DEFAULT_API_KEY = "";
const GEMINI_MODEL = "gemini-3.5-flash";

// Preset test samples
const SAMPLES = {
  netflix: `Dear Customer, your Netflix subscription has expired due to a payment failure. Please update your billing information within 24 hours at http://netflix-billing-update.top/pay to avoid immediate cancellation of your account.`,
  paypal: `PayPal Alert: Unusual sign-in detected from Moscow, Russia. Your account has been restricted for your safety. Confirm your identity and card details at https://paypal-security-verification.org/login to prevent permanent account suspension.`,
  amazon: `Amazon Support Notice: Your order #402-99812-441 for Sony PlayStation 5 ($499.99) has been shipped. If you did NOT place this order, call our 24/7 fraud prevention desk immediately at +1-800-555-0199 or click http://amzn-refund-portal.biz`,
  safe: `Chase: 739201 is your one-time security code. Valid for 10 minutes. Chase will NEVER call or text you asking for this code. If you did not make this request, log into your Chase app or visit chase.com directly.`
};

// State
let activeTab = 'text'; // 'text' | 'image'
let uploadedImage = null; // { dataBase64, mimeType }
let currentReport = null;

// DOM Elements
const textTabPane = document.getElementById('text-tab-pane');
const imageTabPane = document.getElementById('image-tab-pane');
const tabButtons = document.querySelectorAll('.tab-btn');
const messageTextArea = document.getElementById('message-text');
const dropZone = document.getElementById('drop-zone');
const fileInput = document.getElementById('file-input');
const imagePreviewWrapper = document.getElementById('image-preview-wrapper');
const imagePreview = document.getElementById('image-preview');
const dropZonePrompt = document.getElementById('drop-zone-prompt');
const removeImageBtn = document.getElementById('remove-image-btn');
const analyzeBtn = document.getElementById('analyze-btn');
const btnText = analyzeBtn.querySelector('.btn-text');
const btnSpinner = analyzeBtn.querySelector('.spinner');

// States
const emptyState = document.getElementById('empty-state');
const loadingState = document.getElementById('loading-state');
const errorState = document.getElementById('error-state');
const reportContent = document.getElementById('report-content');
const errorTitle = document.getElementById('error-title');
const errorMessage = document.getElementById('error-message');
const retryBtn = document.getElementById('retry-btn');
const resetBtn = document.getElementById('reset-btn');
const copyReportBtn = document.getElementById('copy-report-btn');

// Report elements
const riskScoreVal = document.getElementById('risk-score-val');
const riskLevelBadge = document.getElementById('risk-level-badge');
const claimedEntityVal = document.getElementById('claimed-entity-val');
const warningSignsList = document.getElementById('warning-signs-list');
const safeStepsList = document.getElementById('safe-steps-list');
const scoreCircle = document.getElementById('score-circle');

// Modal Elements
const keyModal = document.getElementById('key-modal');
const toggleKeyBtn = document.getElementById('toggle-key-btn');
const closeModalBtn = document.getElementById('close-modal-btn');
const geminiKeyInput = document.getElementById('gemini-key-input');
const saveKeyBtn = document.getElementById('save-key-btn');
const toggleKeyVisibility = document.getElementById('toggle-key-visibility');
const keyStatusText = document.getElementById('key-status-text');

// Initialize
function init() {
  const savedKey = localStorage.getItem('gemini_api_key');
  if (!savedKey) {
    localStorage.setItem('gemini_api_key', DEFAULT_API_KEY);
  }
  updateKeyStatusDisplay();
  bindEvents();
}

function getApiKey() {
  return localStorage.getItem('gemini_api_key') || DEFAULT_API_KEY;
}

function updateKeyStatusDisplay() {
  const key = getApiKey();
  if (key && key.trim().length > 10) {
    keyStatusText.textContent = "API Key Active ✅";
    geminiKeyInput.value = key;
  } else {
    keyStatusText.textContent = "Set API Key ⚠️";
  }
}

function bindEvents() {
  // Tabs
  tabButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      tabButtons.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      activeTab = btn.dataset.tab;
      if (activeTab === 'text') {
        textTabPane.classList.add('active');
        imageTabPane.classList.remove('active');
      } else {
        textTabPane.classList.remove('active');
        imageTabPane.classList.add('active');
      }
    });
  });

  // Sample chips
  document.querySelectorAll('.preset-chip').forEach(chip => {
    chip.addEventListener('click', () => {
      const sampleKey = chip.dataset.sample;
      if (SAMPLES[sampleKey]) {
        messageTextArea.value = SAMPLES[sampleKey];
      }
    });
  });

  // File dropzone
  dropZone.addEventListener('click', (e) => {
    if (e.target !== removeImageBtn && !uploadedImage) {
      fileInput.click();
    }
  });

  fileInput.addEventListener('change', (e) => {
    const file = e.target.files[0];
    if (file) handleImageFile(file);
  });

  dropZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropZone.classList.add('drag-over');
  });

  dropZone.addEventListener('dragleave', () => {
    dropZone.classList.remove('drag-over');
  });

  dropZone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropZone.classList.remove('drag-over');
    if (e.dataTransfer.files.length > 0) {
      handleImageFile(e.dataTransfer.files[0]);
    }
  });

  // Paste image from clipboard anywhere on page
  window.addEventListener('paste', (e) => {
    const items = (e.clipboardData || e.originalEvent.clipboardData).items;
    for (const item of items) {
      if (item.type.indexOf('image') !== -1) {
        const file = item.getAsFile();
        handleImageFile(file);
        // Switch to image tab automatically
        document.querySelector('[data-tab="image"]').click();
        break;
      }
    }
  });

  removeImageBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    clearImageUpload();
  });

  // Analyze button
  analyzeBtn.addEventListener('click', handleAnalyze);
  retryBtn.addEventListener('click', handleAnalyze);
  resetBtn.addEventListener('click', resetChecker);
  copyReportBtn.addEventListener('click', copyReportToClipboard);

  // Modal events
  toggleKeyBtn.addEventListener('click', () => {
    geminiKeyInput.value = getApiKey();
    keyModal.classList.remove('hidden');
  });

  closeModalBtn.addEventListener('click', () => {
    keyModal.classList.add('hidden');
  });

  saveKeyBtn.addEventListener('click', () => {
    const val = geminiKeyInput.value.trim();
    if (val) {
      localStorage.setItem('gemini_api_key', val);
      updateKeyStatusDisplay();
      keyModal.classList.add('hidden');
    }
  });

  toggleKeyVisibility.addEventListener('click', () => {
    geminiKeyInput.type = geminiKeyInput.type === 'password' ? 'text' : 'password';
  });

  keyModal.addEventListener('click', (e) => {
    if (e.target === keyModal) keyModal.classList.add('hidden');
  });
}

function handleImageFile(file) {
  if (!file.type.startsWith('image/')) {
    alert('Please upload a valid image file (PNG, JPG, WebP).');
    return;
  }

  const reader = new FileReader();
  reader.onload = (e) => {
    const dataUrl = e.target.result;
    const base64Data = dataUrl.split(',')[1];
    uploadedImage = {
      mimeType: file.type,
      base64Data: base64Data,
      dataUrl: dataUrl
    };

    imagePreview.src = dataUrl;
    dropZonePrompt.classList.add('hidden');
    imagePreviewWrapper.classList.remove('hidden');
  };
  reader.readAsDataURL(file);
}

function clearImageUpload() {
  uploadedImage = null;
  fileInput.value = '';
  imagePreview.src = '';
  dropZonePrompt.classList.remove('hidden');
  imagePreviewWrapper.classList.add('hidden');
}

function resetChecker() {
  messageTextArea.value = '';
  clearImageUpload();
  showState('empty');
}

function showState(stateName) {
  emptyState.classList.add('hidden');
  loadingState.classList.add('hidden');
  errorState.classList.add('hidden');
  reportContent.classList.add('hidden');

  if (stateName === 'empty') emptyState.classList.remove('hidden');
  if (stateName === 'loading') loadingState.classList.remove('hidden');
  if (stateName === 'error') errorState.classList.remove('hidden');
  if (stateName === 'report') reportContent.classList.remove('hidden');
}

async function handleAnalyze() {
  const text = messageTextArea.value.trim();

  if (activeTab === 'text' && !text) {
    alert('Please paste a message or choose a sample to analyze.');
    messageTextArea.focus();
    return;
  }

  if (activeTab === 'image' && !uploadedImage) {
    alert('Please drop or select a screenshot image to analyze.');
    return;
  }

  showState('loading');
  setButtonLoading(true);

  try {
    const report = await callGeminiAnalysis({
      type: activeTab,
      text: text,
      image: uploadedImage
    });

    currentReport = report;
    renderReport(report);
    showState('report');
  } catch (err) {
    console.error('Analysis error:', err);
    errorTitle.textContent = "Analysis Error";
    errorMessage.textContent = err.message || "Failed to analyze message with Gemini. Please verify your connection or API key.";
    showState('error');
  } finally {
    setButtonLoading(false);
  }
}

function setButtonLoading(isLoading) {
  if (isLoading) {
    analyzeBtn.disabled = true;
    btnText.textContent = "Inspecting Tactics...";
    btnSpinner.classList.remove('hidden');
  } else {
    analyzeBtn.disabled = false;
    btnText.textContent = "Analyze for Scams 🔍";
    btnSpinner.classList.add('hidden');
  }
}

// Gemini API Call
async function callGeminiAnalysis({ type, text, image }) {
  const apiKey = getApiKey();
  if (!apiKey) {
    throw new Error("Missing Gemini API Key. Click 'Set API Key' in the top right to configure it.");
  }

  const endpoint = `https://generativelanguage.googleapis.com/v1beta/models/${GEMINI_MODEL}:generateContent?key=${apiKey}`;

  const promptText = `You are an expert cybersecurity scam detection assistant.
A user received a message claiming to be from customer support.
Analyze the message or screenshot and identify social engineering tactics (urgency, impersonation, fear, unusual requests for credentials/OTP/gift cards/crypto).

Return valid JSON with:
- risk_score (integer 0-100)
- risk_level ("LOW", "SUSPICIOUS", "HIGH")
- claimed_entity (e.g. "Netflix Support", "PayPal", "Amazon", "Bank", etc.)
- warning_signs (list of specific suspicious elements observed)
- safe_verification_steps (concrete, actionable steps for how the customer can safely verify without clicking links)`;

  let parts = [];

  if (type === 'image' && image) {
    parts.push({
      inline_data: {
        mime_type: image.mimeType,
        data: image.base64Data
      }
    });
    parts.push({
      text: `${promptText}\n\nExamine this screenshot. Extract the text and analyze whether this support message is a scam.`
    });
  } else {
    parts.push({
      text: `${promptText}\n\nAnalyze this customer support message:\n\n${text}`
    });
  }

  const payload = {
    contents: [
      { parts: parts }
    ],
    generationConfig: {
      response_mime_type: "application/json",
      response_schema: {
        type: "OBJECT",
        properties: {
          risk_score: { type: "INTEGER" },
          risk_level: { type: "STRING", enum: ["LOW", "SUSPICIOUS", "HIGH"] },
          claimed_entity: { type: "STRING" },
          warning_signs: {
            type: "ARRAY",
            items: { type: "STRING" }
          },
          safe_verification_steps: {
            type: "ARRAY",
            items: { type: "STRING" }
          }
        },
        required: ["risk_score", "risk_level", "claimed_entity", "warning_signs", "safe_verification_steps"]
      }
    }
  };

  const response = await fetch(endpoint, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  });

  if (!response.ok) {
    const errorJson = await response.json().catch(() => ({}));
    const message = errorJson.error?.message || `HTTP ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  const data = await response.json();
  const rawText = data.candidates?.[0]?.content?.parts?.[0]?.text;

  if (!rawText) {
    throw new Error("No response generated by Gemini model.");
  }

  try {
    return JSON.parse(rawText);
  } catch (err) {
    throw new Error("Received malformed JSON from analysis engine.");
  }
}

// Render Results
function renderReport(report) {
  // Score & Level
  const score = Math.max(0, Math.min(100, report.risk_score || 0));
  const level = (report.risk_level || 'SUSPICIOUS').toUpperCase();

  riskScoreVal.textContent = score;
  claimedEntityVal.textContent = report.claimed_entity || "Unknown";

  riskLevelBadge.className = 'risk-badge';
  scoreCircle.style.boxShadow = '';

  if (level === 'HIGH' || score >= 75) {
    riskLevelBadge.textContent = '🚨 HIGH RISK';
    riskLevelBadge.classList.add('high');
    scoreCircle.style.borderColor = 'var(--danger)';
    scoreCircle.style.boxShadow = '0 0 16px var(--danger-bg)';
  } else if (level === 'SUSPICIOUS' || score >= 40) {
    riskLevelBadge.textContent = '⚠️ SUSPICIOUS';
    riskLevelBadge.classList.add('suspicious');
    scoreCircle.style.borderColor = 'var(--warning)';
    scoreCircle.style.boxShadow = '0 0 16px var(--warning-bg)';
  } else {
    riskLevelBadge.textContent = '✅ LOW RISK';
    riskLevelBadge.classList.add('low');
    scoreCircle.style.borderColor = 'var(--success)';
    scoreCircle.style.boxShadow = '0 0 16px var(--success-bg)';
  }

  // Warning signs
  warningSignsList.innerHTML = '';
  if (!report.warning_signs || report.warning_signs.length === 0) {
    const li = document.createElement('li');
    li.textContent = "No overt red flags detected.";
    warningSignsList.appendChild(li);
  } else {
    report.warning_signs.forEach(sign => {
      const li = document.createElement('li');
      li.textContent = sign;
      warningSignsList.appendChild(li);
    });
  }

  // Safe verification steps
  safeStepsList.innerHTML = '';
  if (!report.safe_verification_steps || report.safe_verification_steps.length === 0) {
    const li = document.createElement('li');
    li.textContent = "Navigate directly to the official website or verified mobile app.";
    safeStepsList.appendChild(li);
  } else {
    report.safe_verification_steps.forEach(step => {
      const li = document.createElement('li');
      li.textContent = step;
      safeStepsList.appendChild(li);
    });
  }
}

// Copy report to clipboard
function copyReportToClipboard() {
  if (!currentReport) return;

  const emoji = currentReport.risk_score >= 75 ? '🚨' : (currentReport.risk_score >= 40 ? '⚠️' : '✅');
  let text = `${emoji} Scam-O-Not Analysis Report\n`;
  text += `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n`;
  text += `Risk Level: ${currentReport.risk_level} (${currentReport.risk_score}/100)\n`;
  text += `Claimed Entity: ${currentReport.claimed_entity}\n\n`;

  text += `🚩 Observed Warning Signs:\n`;
  (currentReport.warning_signs || []).forEach(sign => {
    text += `• ${sign}\n`;
  });

  text += `\n🛡️ Safe Verification Steps:\n`;
  (currentReport.safe_verification_steps || []).forEach((step, i) => {
    text += `${i + 1}. ${step}\n`;
  });

  navigator.clipboard.writeText(text).then(() => {
    const originalText = copyReportBtn.textContent;
    copyReportBtn.textContent = "✅ Copied!";
    setTimeout(() => {
      copyReportBtn.textContent = originalText;
    }, 2000);
  }).catch(() => {
    alert("Copied report to clipboard!");
  });
}

// Kickoff
window.addEventListener('DOMContentLoaded', init);
