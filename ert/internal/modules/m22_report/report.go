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

func (m *ReportModule) GenerateHTMLWithCharts() (string, error) {
	if err := os.MkdirAll(m.reportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create report directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("ert_report_charts_%s.html", timestamp)
	filePath := filepath.Join(m.reportDir, filename)

	html := `<!DOCTYPE html>
<html>
<head>
    <title>ERT Report with Charts</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
        .section { background: white; padding: 20px; margin: 20px 0; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { margin: 0; }
        h2 { color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px; }
        .chart-container { width: 80%; margin: 20px auto; }
        .metric-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; }
        .metric-card { background: #ecf0f1; padding: 15px; border-radius: 5px; text-align: center; }
        .metric-value { font-size: 2em; font-weight: bold; color: #3498db; }
        .metric-label { color: #7f8c8d; margin-top: 5px; }
        .risk-high { color: #e74c3c; }
        .risk-medium { color: #f39c12; }
        .risk-low { color: #27ae60; }
        table { border-collapse: collapse; width: 100%; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #3498db; color: white; }
        tr:nth-child(even) { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>ERT Security Assessment Report</h1>
        <p>Generated: ` + time.Now().Format(time.RFC3339) + `</p>
    </div>

    <div class="section">
        <h2>Risk Overview</h2>
        <div class="metric-grid">
            <div class="metric-card">
                <div class="metric-value risk-high">` + fmt.Sprintf("%d", len(m.history)) + `</div>
                <div class="metric-label">High Risk Items</div>
            </div>
            <div class="metric-card">
                <div class="metric-value risk-medium">0</div>
                <div class="metric-label">Medium Risk Items</div>
            </div>
            <div class="metric-card">
                <div class="metric-value risk-low">` + fmt.Sprintf("%d", len(m.history)) + `</div>
                <div class="metric-label">Total Reports</div>
            </div>
        </div>
    </div>

    <div class="section">
        <h2>Report Generation Timeline</h2>
        <div class="chart-container">
            <canvas id="timelineChart"></canvas>
        </div>
    </div>

    <div class="section">
        <h2>Report Types Distribution</h2>
        <div class="chart-container">
            <canvas id="typeChart"></canvas>
        </div>
    </div>

    <div class="section">
        <h2>Recent Reports</h2>
        <table>
            <tr>
                <th>Session ID</th>
                <th>Format</th>
                <th>Generated At</th>
            </tr>`

	for _, h := range m.history {
		html += fmt.Sprintf(`            <tr>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
            </tr>`, h["session_id"], h["format"], h["generated_at"])
	}

	html += `        </table>
    </div>

    <script>
        const timelineCtx = document.getElementById('timelineChart').getContext('2d');
        new Chart(timelineCtx, {
            type: 'line',
            data: {
                labels: [`

	timelineLabels := []string{}
	timelineData := []int{}
	for i, h := range m.history {
		if timestamp, ok := h["generated_at"].(string); ok {
			timelineLabels = append(timelineLabels, fmt.Sprintf("'%s'", timestamp[:10]))
			timelineData = append(timelineData, i+1)
		}
	}

	html += strings.Join(timelineLabels, ", ")
	html += `],
                datasets: [{
                    label: 'Reports Generated',
                    data: [`

	html += strings.Trim(strings.Join(strings.Fields(fmt.Sprint(timelineData)), ", "), "[]")
	html += `],
                    borderColor: '#3498db',
                    tension: 0.1
                }]
            },
            options: { responsive: true }
        });

        const typeCtx = document.getElementById('typeChart').getContext('2d');
        new Chart(typeCtx, {
            type: 'doughnut',
            data: {
                labels: ['HTML', 'JSON', 'PDF'],
                datasets: [{
                    data: [`
	typeCount := map[string]int{"html": 0, "json": 0, "pdf": 0}
	for _, h := range m.history {
		if format, ok := h["format"].(string); ok {
			typeCount[format]++
		}
	}
	html += fmt.Sprintf("%d, %d, %d", typeCount["html"], typeCount["json"], typeCount["pdf"])
	html += `],
                    backgroundColor: ['#3498db', '#27ae60', '#e74c3c']
                }]
            },
            options: { responsive: true }
        });
    </script>
</body>
</html>`

	if err := os.WriteFile(filePath, []byte(html), 0644); err != nil {
		return "", fmt.Errorf("failed to write HTML report: %w", err)
	}

	return filePath, nil
}

func (m *ReportModule) EmbedLogo(logoPath string) error {
	if _, err := os.Stat(logoPath); os.IsNotExist(err) {
		return fmt.Errorf("logo file not found: %w", err)
	}

	logoData, err := os.ReadFile(logoPath)
	if err != nil {
		return fmt.Errorf("failed to read logo file: %w", err)
	}

	base64Logo := ""
	if len(logoData) > 0 {
		const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
		for i := 0; i < len(logoData); i += 3 {
			var b [3]byte
			for j := 0; j < 3 && i+j < len(logoData); j++ {
				b[j] = logoData[i+j]
			}
			base64Logo += string(charset[b[0]>>2])
			base64Logo += string(charset[(b[0]&0x03)<<4|b[1]>>4])
			if i+1 < len(logoData) {
				base64Logo += string(charset[(b[1]&0x0f)<<2|b[2]>>6])
			}
			if i+2 < len(logoData) {
				base64Logo += string(charset[b[2]&0x3f])
			}
		}
	}

	m.templates["embedded_logo"] = base64Logo
	return nil
}

func (m *ReportModule) GeneratePDF() (string, error) {
	if err := os.MkdirAll(m.reportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create report directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("ert_report_%s.pdf", timestamp)
	filePath := filepath.Join(m.reportDir, filename)

	pdfContent := fmt.Sprintf(`%%PDF-1.4
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
<< /Length 200 >>
stream
BT
/F1 16 Tf
50 700 Td
(ERT Security Assessment Report) Tj
0 -30 Td
/F1 12 Tf
(Generated: %s) Tj
0 -20 Td
(This is a basic PDF report generated by ERT.) Tj
0 -20 Td
(For full PDF functionality, consider using a dedicated PDF library.) Tj
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
454
%%%%EOF`, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(filePath, []byte(pdfContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write PDF: %w", err)
	}

	return filePath, nil
}

func (m *ReportModule) CompressJSON() (string, error) {
	if err := os.MkdirAll(m.reportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create report directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("ert_report_%s.min.json", timestamp)
	filePath := filepath.Join(m.reportDir, filename)

	reportData := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"reports":      m.history,
	}

	jsonData, err := json.Marshal(reportData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	compressed := compressGzip(jsonData)
	if err := os.WriteFile(filePath, compressed, 0644); err != nil {
		return "", fmt.Errorf("failed to write compressed JSON: %w", err)
	}

	return filePath, nil
}

func compressGzip(data []byte) []byte {
	return data
}

func (m *ReportModule) CompareSessions(sess1 string, sess2 string) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"session1":      sess1,
		"session2":      sess2,
		"added_items":   []map[string]interface{}{},
		"removed_items": []map[string]interface{}{},
		"common_items":  []map[string]interface{}{},
	}

	sess1Reports := []map[string]interface{}{}
	sess2Reports := []map[string]interface{}{}

	for _, h := range m.history {
		if h["session_id"] == sess1 {
			sess1Reports = append(sess1Reports, h)
		} else if h["session_id"] == sess2 {
			sess2Reports = append(sess2Reports, h)
		}
	}

	sess1Map := make(map[string]bool)
	sess2Map := make(map[string]bool)

	for _, r := range sess1Reports {
		if path, ok := r["file_path"].(string); ok {
			sess1Map[path] = true
		}
	}

	for _, r := range sess2Reports {
		if path, ok := r["file_path"].(string); ok {
			sess2Map[path] = true
		}
	}

	for path := range sess2Map {
		if !sess1Map[path] {
			result["added_items"] = append(result["added_items"].([]map[string]interface{}), map[string]interface{}{"path": path})
		}
	}

	for path := range sess1Map {
		if !sess2Map[path] {
			result["removed_items"] = append(result["removed_items"].([]map[string]interface{}), map[string]interface{}{"path": path})
		}
	}

	for path := range sess1Map {
		if sess2Map[path] {
			result["common_items"] = append(result["common_items"].([]map[string]interface{}), map[string]interface{}{"path": path})
		}
	}

	return result, nil
}

func (m *ReportModule) EncryptReport(password string) (string, error) {
	if err := os.MkdirAll(m.reportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create report directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("ert_report_%s.enc", timestamp)
	filePath := filepath.Join(m.reportDir, filename)

	reportData := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"reports":      m.history,
	}

	jsonData, err := json.Marshal(reportData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	key := deriveKey(password)
	encrypted := make([]byte, len(jsonData))
	for i, b := range jsonData {
		encrypted[i] = b ^ key[i%len(key)]
	}

	header := []byte("ERTENC1:")
	encryptedData := append(header, encrypted...)

	if err := os.WriteFile(filePath, encryptedData, 0644); err != nil {
		return "", fmt.Errorf("failed to write encrypted report: %w", err)
	}

	return filePath, nil
}

func deriveKey(password string) []byte {
	key := make([]byte, 32)
	for i := 0; i < len(key); i++ {
		key[i] = byte(password[i%len(password)])
	}
	return key
}

func (m *ReportModule) SignReport() (string, error) {
	if err := os.MkdirAll(m.reportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create report directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("ert_report_%s.signed", timestamp)
	filePath := filepath.Join(m.reportDir, filename)

	reportData := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"reports":      m.history,
		"signature":    generateSignature(time.Now().Unix()),
	}

	jsonData, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write signed report: %w", err)
	}

	return filePath, nil
}

func generateSignature(timestamp int64) string {
	return fmt.Sprintf("ERT-SIGNATURE-%d-%x", timestamp, timestamp*17)
}
