package analyzer

import (
	"context"
	"fmt"
	"regexp"

	webrisk "cloud.google.com/go/webrisk/apiv1"
	webriskpb "cloud.google.com/go/webrisk/apiv1/webriskpb"
)

// urlPattern is a simple regex to find URLs inside arbitrary text.
var urlPattern = regexp.MustCompile(`https?://[^\s"'<>]+`)

// WebRiskChecker queries the Google Web Risk API for known malicious URLs.
type WebRiskChecker struct {
	client *webrisk.WebRiskClient
}

// NewWebRiskChecker creates a WebRiskChecker using Application Default Credentials.
func NewWebRiskChecker(ctx context.Context) (*WebRiskChecker, error) {
	c, err := webrisk.NewWebRiskClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("webrisk client: %w", err)
	}
	return &WebRiskChecker{client: c}, nil
}

// Close releases resources held by the client.
func (w *WebRiskChecker) Close() {
	w.client.Close()
}

// ThreatResult describes whether a URL is flagged as a known threat.
type ThreatResult struct {
	URL         string
	IsThreat    bool
	ThreatTypes []string
}

// CheckURLs extracts URLs from the given text and checks each against Web Risk.
// It returns one ThreatResult per URL found (may be empty if no URLs are present).
func (w *WebRiskChecker) CheckURLs(ctx context.Context, text string) ([]ThreatResult, error) {
	urls := urlPattern.FindAllString(text, -1)
	if len(urls) == 0 {
		return nil, nil
	}

	threatTypes := []webriskpb.ThreatType{
		webriskpb.ThreatType_MALWARE,
		webriskpb.ThreatType_SOCIAL_ENGINEERING,
		webriskpb.ThreatType_UNWANTED_SOFTWARE,
	}

	var results []ThreatResult
	for _, u := range urls {
		req := &webriskpb.SearchUrisRequest{
			Uri:         u,
			ThreatTypes: threatTypes,
		}
		resp, err := w.client.SearchUris(ctx, req)
		if err != nil {
			// Treat API errors as non-fatal; log and continue.
			results = append(results, ThreatResult{URL: u, IsThreat: false})
			continue
		}

		var found []string
		if resp.Threat != nil {
			for _, tt := range resp.Threat.ThreatTypes {
				found = append(found, tt.String())
			}
		}
		results = append(results, ThreatResult{
			URL:         u,
			IsThreat:    len(found) > 0,
			ThreatTypes: found,
		})
	}
	return results, nil
}

// AppendWebRiskWarnings adds Web Risk findings as extra warning signs to the report.
// Call this after Gemini analysis to enrich the report with confirmed URL threats.
func AppendWebRiskWarnings(warningSigns []string, threats []ThreatResult) []string {
	for _, t := range threats {
		if t.IsThreat {
			warningSigns = append(warningSigns,
				fmt.Sprintf("URL flagged by Google Web Risk as %v: %s", t.ThreatTypes, t.URL),
			)
		}
	}
	return warningSigns
}
