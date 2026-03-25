//go:build windows

package m25_codec

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/yourname/ert/internal/registry"
)

type CodecModule struct {
	ctx           context.Context
	storage       registry.Storage
	history       []map[string]interface{}
	enableHistory bool
	maxHistory    int
}

func New() *CodecModule {
	return &CodecModule{
		history:       []map[string]interface{}{},
		enableHistory: true,
		maxHistory:    100,
	}
}

func (m *CodecModule) ID() int       { return 25 }
func (m *CodecModule) Name() string  { return "codec" }
func (m *CodecModule) Priority() int { return 2 }

func (m *CodecModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *CodecModule) Collect(ctx context.Context) error {
	m.history = []map[string]interface{}{}
	return nil
}

func (m *CodecModule) Stop() error {
	return nil
}

func (m *CodecModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.history))
	for _, h := range m.history {
		result = append(result, h)
	}
	return result, nil
}

func (m *CodecModule) Encode(input string, codecType string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("input cannot be empty")
	}

	var result string
	var err error

	switch strings.ToLower(codecType) {
	case "base64":
		result = m.encodeBase64(input, false)
	case "base64url":
		result = m.encodeBase64(input, true)
	case "hex":
		result = m.encodeHex(input)
	case "unicode":
		result = m.encodeUnicode(input)
	case "url":
		result = m.encodeURL(input)
	case "html":
		result = m.encodeHTML(input)
	case "binary":
		result = m.encodeBinary(input)
	case "octal":
		result = m.encodeOctal(input)
	default:
		return "", fmt.Errorf("unsupported codec type: %s", codecType)
	}

	if err == nil {
		m.addToHistory("encode", codecType, input, result, "success")
	}

	return result, err
}

func (m *CodecModule) Decode(input string, codecType string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("input cannot be empty")
	}

	var result string
	var err error

	switch strings.ToLower(codecType) {
	case "base64":
		result, err = m.decodeBase64(input, false)
	case "base64url":
		result, err = m.decodeBase64(input, true)
	case "hex":
		result, err = m.decodeHex(input)
	case "unicode":
		result, err = m.decodeUnicode(input)
	case "url":
		result, err = m.decodeURL(input)
	case "html":
		result, err = m.decodeHTML(input)
	case "binary":
		result, err = m.decodeBinary(input)
	case "octal":
		result, err = m.decodeOctal(input)
	default:
		return "", fmt.Errorf("unsupported codec type: %s", codecType)
	}

	if err == nil {
		m.addToHistory("decode", codecType, input, result, "success")
	}

	return result, err
}

func (m *CodecModule) AutoDetect(input string) ([]map[string]interface{}, error) {
	if input == "" {
		return nil, fmt.Errorf("input cannot be empty")
	}

	results := []map[string]interface{}{}

	results = append(results, m.detectBase64(input)...)
	results = append(results, m.detectHex(input)...)
	results = append(results, m.detectUnicode(input)...)
	results = append(results, m.detectURL(input)...)
	results = append(results, m.detectHTML(input)...)
	results = append(results, m.detectBinary(input)...)

	if len(results) == 0 {
		results = append(results, map[string]interface{}{
			"type":     "unknown",
			"input":    input,
			"message":  "Could not determine encoding type",
			"detected": false,
		})
	}

	return results, nil
}

func (m *CodecModule) encodeBase64(input string, urlSafe bool) string {
	if urlSafe {
		return base64.URLEncoding.EncodeToString([]byte(input))
	}
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func (m *CodecModule) decodeBase64(input string, urlSafe bool) (string, error) {
	var data []byte
	var err error

	if urlSafe {
		data, err = base64.URLEncoding.DecodeString(input)
	} else {
		data, err = base64.StdEncoding.DecodeString(input)
	}

	if err != nil {
		return "", fmt.Errorf("invalid base64 input: %w", err)
	}
	return string(data), nil
}

func (m *CodecModule) encodeHex(input string) string {
	return hex.EncodeToString([]byte(input))
}

func (m *CodecModule) decodeHex(input string) (string, error) {
	data, err := hex.DecodeString(input)
	if err != nil {
		return "", fmt.Errorf("invalid hex input: %w", err)
	}
	return string(data), nil
}

func (m *CodecModule) encodeUnicode(input string) string {
	var result strings.Builder
	for _, r := range input {
		result.WriteString(fmt.Sprintf("\\u%04x", r))
	}
	return result.String()
}

func (m *CodecModule) decodeUnicode(input string) (string, error) {
	pattern := regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)
	matches := pattern.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		pattern = regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)
		if !pattern.MatchString(input) {
			return "", fmt.Errorf("invalid unicode escape sequence")
		}
	}

	result := pattern.ReplaceAllStringFunc(input, func(match string) string {
		hexVal := match[2:]
		var codePoint uint32
		fmt.Sscanf(hexVal, "%x", &codePoint)
		return string(rune(codePoint))
	})

	return result, nil
}

func (m *CodecModule) encodeURL(input string) string {
	return url.QueryEscape(input)
}

func (m *CodecModule) decodeURL(input string) (string, error) {
	decoded, err := url.QueryUnescape(input)
	if err != nil {
		return "", fmt.Errorf("invalid URL encoding: %w", err)
	}
	return decoded, nil
}

func (m *CodecModule) encodeHTML(input string) string {
	var result strings.Builder
	for _, r := range input {
		switch r {
		case '&':
			result.WriteString("&amp;")
		case '<':
			result.WriteString("&lt;")
		case '>':
			result.WriteString("&gt;")
		case '"':
			result.WriteString("&quot;")
		case '\'':
			result.WriteString("&#39;")
		default:
			if r < 128 {
				result.WriteRune(r)
			} else {
				result.WriteString(fmt.Sprintf("&#%d;", r))
			}
		}
	}
	return result.String()
}

func (m *CodecModule) decodeHTML(input string) (string, error) {
	pattern := regexp.MustCompile(`&#(\d+);`)
	result := pattern.ReplaceAllStringFunc(input, func(match string) string {
		var code int
		fmt.Sscanf(match[2:len(match)-1], "%d", &code)
		return string(rune(code))
	})

	htmlEntities := map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
		"&apos;": "'",
		"&nbsp;": " ",
	}

	for entity, char := range htmlEntities {
		result = strings.ReplaceAll(result, entity, char)
	}

	return result, nil
}

func (m *CodecModule) encodeBinary(input string) string {
	var result strings.Builder
	for _, r := range input {
		binary := ""
		for i := 7; i >= 0; i-- {
			if (r>>i)&1 == 1 {
				binary += "1"
			} else {
				binary += "0"
			}
		}
		result.WriteString(binary)
		result.WriteString(" ")
	}
	return strings.TrimSpace(result.String())
}

func (m *CodecModule) decodeBinary(input string) (string, error) {
	input = strings.ReplaceAll(input, " ", "")
	input = strings.ReplaceAll(input, "\n", "")

	if len(input)%8 != 0 {
		return "", fmt.Errorf("invalid binary input: length must be multiple of 8")
	}

	var result strings.Builder
	for i := 0; i < len(input); i += 8 {
		byteStr := input[i : i+8]
		var b byte
		for _, c := range byteStr {
			if c != '0' && c != '1' {
				return "", fmt.Errorf("invalid binary character: %c", c)
			}
			b = b<<1 | byte(c-'0')
		}
		result.WriteByte(b)
	}

	return result.String(), nil
}

func (m *CodecModule) encodeOctal(input string) string {
	var result strings.Builder
	for _, r := range input {
		result.WriteString(fmt.Sprintf("%03o", r))
		result.WriteString(" ")
	}
	return strings.TrimSpace(result.String())
}

func (m *CodecModule) decodeOctal(input string) (string, error) {
	input = strings.ReplaceAll(input, " ", "")
	input = strings.ReplaceAll(input, "\n", "")
	input = strings.TrimSpace(input)

	if len(input)%3 != 0 {
		return "", fmt.Errorf("invalid octal input: length must be multiple of 3")
	}

	var result strings.Builder
	for i := 0; i < len(input); i += 3 {
		octalStr := input[i : i+3]
		var val byte
		for _, c := range octalStr {
			if c < '0' || c > '7' {
				return "", fmt.Errorf("invalid octal character: %c", c)
			}
			val = val<<3 | byte(c-'0')
		}
		result.WriteByte(val)
	}

	return result.String(), nil
}

func (m *CodecModule) detectBase64(input string) []map[string]interface{} {
	results := []map[string]interface{}{}

	pattern := regexp.MustCompile(`^[A-Za-z0-9+/]+=*$`)
	if pattern.MatchString(input) && len(input) >= 4 {
		tryDecode, err := m.decodeBase64(input, false)
		if err == nil && tryDecode != "" && isPrintable(tryDecode) {
			results = append(results, map[string]interface{}{
				"type":       "base64",
				"input":      input,
				"decoded":    tryDecode,
				"confidence": 0.8,
				"detected":   true,
			})
		}
	}

	patternURL := regexp.MustCompile(`^[A-Za-z0-9_-]+=*$`)
	if patternURL.MatchString(input) && len(input) >= 4 {
		tryDecode, err := m.decodeBase64(input, true)
		if err == nil && tryDecode != "" && isPrintable(tryDecode) {
			results = append(results, map[string]interface{}{
				"type":       "base64url",
				"input":      input,
				"decoded":    tryDecode,
				"confidence": 0.7,
				"detected":   true,
			})
		}
	}

	return results
}

func (m *CodecModule) detectHex(input string) []map[string]interface{} {
	results := []map[string]interface{}{}

	pattern := regexp.MustCompile(`^[0-9A-Fa-f]+$`)
	if pattern.MatchString(input) && len(input) >= 2 && len(input)%2 == 0 {
		tryDecode, err := m.decodeHex(input)
		if err == nil && tryDecode != "" && isPrintable(tryDecode) {
			results = append(results, map[string]interface{}{
				"type":       "hex",
				"input":      input,
				"decoded":    tryDecode,
				"confidence": 0.9,
				"detected":   true,
			})
		}
	}

	return results
}

func (m *CodecModule) detectUnicode(input string) []map[string]interface{} {
	results := []map[string]interface{}{}

	pattern := regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	if pattern.MatchString(input) {
		tryDecode, err := m.decodeUnicode(input)
		if err == nil && tryDecode != "" {
			results = append(results, map[string]interface{}{
				"type":       "unicode",
				"input":      input,
				"decoded":    tryDecode,
				"confidence": 0.95,
				"detected":   true,
			})
		}
	}

	return results
}

func (m *CodecModule) detectURL(input string) []map[string]interface{} {
	results := []map[string]interface{}{}

	if strings.Contains(input, "%") && strings.Contains(input, "%2") {
		tryDecode, err := m.decodeURL(input)
		if err == nil && tryDecode != input {
			results = append(results, map[string]interface{}{
				"type":       "url",
				"input":      input,
				"decoded":    tryDecode,
				"confidence": 0.9,
				"detected":   true,
			})
		}
	}

	return results
}

func (m *CodecModule) detectHTML(input string) []map[string]interface{} {
	results := []map[string]interface{}{}

	pattern := regexp.MustCompile(`&[a-zA-Z]+;|&#\d+;`)
	if pattern.MatchString(input) {
		tryDecode, err := m.decodeHTML(input)
		if err == nil && tryDecode != input {
			results = append(results, map[string]interface{}{
				"type":       "html",
				"input":      input,
				"decoded":    tryDecode,
				"confidence": 0.85,
				"detected":   true,
			})
		}
	}

	return results
}

func (m *CodecModule) detectBinary(input string) []map[string]interface{} {
	results := []map[string]interface{}{}

	cleaned := strings.ReplaceAll(input, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "\n", "")

	pattern := regexp.MustCompile(`^[01]+$`)
	if pattern.MatchString(cleaned) && len(cleaned) >= 8 && len(cleaned)%8 == 0 {
		tryDecode, err := m.decodeBinary(cleaned)
		if err == nil && tryDecode != "" && isPrintable(tryDecode) {
			results = append(results, map[string]interface{}{
				"type":       "binary",
				"input":      input,
				"decoded":    tryDecode,
				"confidence": 0.9,
				"detected":   true,
			})
		}
	}

	return results
}

func (m *CodecModule) addToHistory(op, codec, input, output, status string) {
	if !m.enableHistory {
		return
	}

	entry := map[string]interface{}{
		"operation": op,
		"codec":     codec,
		"input":     input,
		"output":    output,
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	m.history = append(m.history, entry)

	if len(m.history) > m.maxHistory {
		m.history = m.history[len(m.history)-m.maxHistory:]
	}
}

func (m *CodecModule) GetHistory() []map[string]interface{} {
	return m.history
}

func (m *CodecModule) ClearHistory() {
	m.history = []map[string]interface{}{}
}

func (m *CodecModule) SetHistoryEnabled(enabled bool, maxSize int) {
	m.enableHistory = enabled
	m.maxHistory = maxSize
}

func isPrintable(s string) bool {
	for _, r := range s {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
		if r > 126 {
			return false
		}
	}
	return true
}

func BatchEncode(input string, codecTypes []string) (map[string]string, error) {
	module := New()
	results := make(map[string]string)

	for _, codec := range codecTypes {
		result, err := module.Encode(input, codec)
		if err != nil {
			results[codec] = fmt.Sprintf("error: %v", err)
		} else {
			results[codec] = result
		}
	}

	return results, nil
}

func BatchDecode(input string, codecTypes []string) (map[string]string, error) {
	module := New()
	results := make(map[string]string)

	for _, codec := range codecTypes {
		result, err := module.Decode(input, codec)
		if err != nil {
			results[codec] = fmt.Sprintf("error: %v", err)
		} else {
			results[codec] = result
		}
	}

	return results, nil
}
