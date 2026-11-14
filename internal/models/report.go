package models

// AnalysisReport holds the structured output from Gemini scam analysis.
type AnalysisReport struct {
	RiskScore             int      `json:"risk_score"`
	RiskLevel             string   `json:"risk_level"`
	ClaimedEntity         string   `json:"claimed_entity"`
	WarningSigns          []string `json:"warning_signs"`
	SafeVerificationSteps []string `json:"safe_verification_steps"`
}
