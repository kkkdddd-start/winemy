//go:build windows

package m22_report

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourname/ert/internal/registry"
)

type ReportModule struct {
	ctx       context.Context
	storage   registry.Storage
	reportDir string
	templates map[string]string
	history   []map[string]interface{}
}

func New() *ReportModule {
	return &ReportModule{
		reportDir: "./data/reports",
		templates: map[string]string{},
		history:   []map[string]interface{}{},
	}
}

func (m *ReportModule) ID() int       { return 22 }
func (m *ReportModule) Name() string  { return "report" }
func (m *ReportModule) Priority() int { return 2 }

func (m *ReportModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	if err := os.MkdirAll(m.reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}
	m.loadTemplates()
	return nil
}

func (m *ReportModule) loadTemplates() {
	m.templates = map[string]string{
		"html": `<!DOCTYPE html>
<html>
<head>
    <title>ERT Report - {{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        .section { margin: 20px 0; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
        .risk-high { color: red; font-weight: bold; }
        .risk-medium { color: orange; font-weight: bold; }
        .risk-low { color: green; }
    </style>
</head>
<body>
    <h1>ERT Report</h1>
    <p>Generated: {{.GeneratedAt}}</p>
    <p>Session: {{.SessionID}}</p>
    {{range .Sections}}
    <div class="section">
        <h2>{{.Title}}</h2>
        {{.Content}}
    </div>
    {{end}}
</body>
</html>`,
		"json": `{"report": {"title": "{{.Title}}", "generated_at": "{{.GeneratedAt}}", "session_id": "{{.SessionID}}"}}`,
	}
}

func (m *ReportModule) Collect(ctx context.Context) error {
	m.history = []map[string]interface{}{}
	return nil
}

func (m *ReportModule) Stop() error {
	return nil
}

func (m *ReportModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.history))
	for _, h := range m.history {
		result = append(result, h)
	}
	return result, nil
}

func (m *ReportModule) ExportReport(format string, sessionID string) (string, error) {
	if err := os.MkdirAll(m.reportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create report directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("ert_report_%s_%s.%s", sessionID, timestamp, format)
	filePath := filepath.Join(m.reportDir, filename)

	var content string
	switch format {
	case "html":
		content = m.generateHTMLReport(sessionID)
	case "json":
		content = m.generateJSONReport(sessionID)
	case "pdf":
		content = m.generatePDFReport(sessionID)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write report file: %w", err)
	}

	entry := map[string]interface{}{
		"session_id":   sessionID,
		"format":       format,
		"file_path":    filePath,
		"generated_at": time.Now().Format(time.RFC3339),
	}
	m.history = append(m.history, entry)

	return filePath, nil
}

func (m *ReportModule) generateHTMLReport(sessionID string) string {
	data := map[string]interface{}{
		"Title":       fmt.Sprintf("ERT Report - Session %s", sessionID),
		"GeneratedAt": time.Now().Format(time.RFC3339),
		"SessionID":   sessionID,
		"Sections": []map[string]interface{}{
			{"Title": "System Information", "Content": "<p>System data collected during the session.</p>"},
			{"Title": "Process Analysis", "Content": "<p>Process information collected during the session.</p>"},
			{"Title": "Network Connections", "Content": "<p>Network connection data collected during the session.</p>"},
		},
	}

	template := m.templates["html"]
	for key, value := range data {
		template = strings.ReplaceAll(template, fmt.Sprintf("{{.%s}}", key), fmt.Sprintf("%v", value))
	}

	sectionsStr := ""
	for _, section := range data["Sections"].([]map[string]interface{}) {
		sectionsStr += fmt.Sprintf(`    <div class="section">
        <h2>%s</h2>
        %s
    </div>
`, section["Title"], section["Content"])
	}
	template = strings.ReplaceAll(template, "{{range .Sections}}\n    <div class=\"section\">\n        <h2>{{.Title}}</h2>\n        {{.Content}}\n    </div>\n    {{end}}", sectionsStr)

	return template
}

func (m *ReportModule) generateJSONReport(sessionID string) string {
	report := map[string]interface{}{
		"report": map[string]interface{}{
			"title":        fmt.Sprintf("ERT Report - Session %s", sessionID),
			"generated_at": time.Now().Format(time.RFC3339),
			"session_id":   sessionID,
			"sections": []map[string]interface{}{
				{"title": "System Information", "content": "System data collected during the session."},
				{"title": "Process Analysis", "content": "Process information collected during the session."},
				{"title": "Network Connections", "content": "Network connection data collected during the session."},
			},
		},
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to generate json report: %v"}`, err)
	}

	return string(jsonData)
}

func (m *ReportModule) generatePDFReport(sessionID string) string {
	return fmt.Sprintf(`%%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 100 >>
stream
BT
/F1 12 Tf
50 700 Td
(ERT Report - Session %s) Tj
0 -20 Td
(Generated: %s) Tj
0 -20 Td
(This is a placeholder PDF report.) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000214 00000 n 
trailer
<< /Size 5 /Root 1 0 R >>
startxref
354
%%%%EOF`, sessionID, time.Now().Format(time.RFC3339))
}

func (m *ReportModule) ListReports() []map[string]interface{} {
	return m.history
}

func (m *ReportModule) DeleteReport(filePath string) error {
	for i, h := range m.history {
		if h["file_path"] == filePath {
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete report file: %w", err)
			}
			m.history = append(m.history[:i], m.history[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("report not found in history")
}

func (m *ReportModule) GetReportPath() string {
	return m.reportDir
}
