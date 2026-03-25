//go:build windows

package codec

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type CodecEngine struct {
	historyDB     interface{}
	maxHistory    int
	enableHistory bool
}

type CodecType string

const (
	CodecBase64    CodecType = "base64"
	CodecBase64URL CodecType = "base64url"
	CodecHex       CodecType = "hex"
	CodecUnicode   CodecType = "unicode"
	CodecURL       CodecType = "url"
	CodecHTML      CodecType = "html"
	CodecBinary    CodecType = "binary"
)

func New() *CodecEngine {
	return &CodecEngine{
		maxHistory:    100,
		enableHistory: true,
	}
}

func (e *CodecEngine) Encode(input string, codecType CodecType) (string, error) {
	switch codecType {
	case CodecBase64:
		return base64.StdEncoding.EncodeToString([]byte(input)), nil
	case CodecBase64URL:
		return base64.URLEncoding.EncodeToString([]byte(input)), nil
	case CodecHex:
		return hex.EncodeToString([]byte(input)), nil
	case CodecUnicode:
		return encodeUnicode(input), nil
	case CodecURL:
		return url.QueryEscape(input), nil
	case CodecHTML:
		return encodeHTML(input), nil
	case CodecBinary:
		return encodeBinary(input), nil
	default:
		return "", fmt.Errorf("unsupported codec type: %s", codecType)
	}
}

func (e *CodecEngine) Decode(input string, codecType CodecType) (string, error) {
	switch codecType {
	case CodecBase64:
		data, err := base64.StdEncoding.DecodeString(input)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case CodecBase64URL:
		data, err := base64.URLEncoding.DecodeString(input)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case CodecHex:
		data, err := hex.DecodeString(input)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case CodecUnicode:
		return decodeUnicode(input)
	case CodecURL:
		return url.QueryUnescape(input)
	case CodecHTML:
		return decodeHTML(input)
	case CodecBinary:
		return decodeBinary(input)
	default:
		return "", fmt.Errorf("unsupported codec type: %s", codecType)
	}
}

func (e *CodecEngine) AutoDetect(input string) ([]map[string]interface{}, error) {
	results := []map[string]interface{}{}

	codecs := []CodecType{
		CodecBase64, CodecBase64URL, CodecHex, CodecUnicode, CodecURL, CodecHTML,
	}

	for _, ct := range codecs {
		decoded, err := e.Decode(input, ct)
		if err == nil && len(decoded) > 0 && len(decoded) < 1000 {
			results = append(results, map[string]interface{}{
				"codec_type": ct,
				"input":      input,
				"output":     decoded,
				"operation":  "decode",
			})
		}
	}

	return results, nil
}

func encodeUnicode(s string) string {
	var builder strings.Builder
	for _, r := range s {
		fmt.Fprintf(&builder, "\\u%04x", r)
	}
	return builder.String()
}

func decodeUnicode(s string) (string, error) {
	re := regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		hex := match[2:]
		var r rune
		fmt.Sscanf(hex, "%x", &r)
		return string(r)
	}), nil
}

func encodeHTML(s string) string {
	var builder strings.Builder
	for _, r := range s {
		fmt.Fprintf(&builder, "&#%d;", r)
	}
	return builder.String()
}

func decodeHTML(s string) (string, error) {
	re := regexp.MustCompile(`&#(\d+);`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		var num int
		fmt.Sscanf(match[2:len(match)-1], "%d", &num)
		return string(rune(num))
	}), nil
}

func encodeBinary(s string) string {
	var builder strings.Builder
	for _, b := range []byte(s) {
		fmt.Fprintf(&builder, "%08b", b)
	}
	return builder.String()
}

func decodeBinary(s string) (string, error) {
	s = strings.ReplaceAll(s, " ", "")
	if len(s)%8 != 0 {
		return "", fmt.Errorf("invalid binary string length")
	}

	var builder strings.Builder
	for i := 0; i < len(s); i += 8 {
		var b byte
		_, err := fmt.Sscanf(s[i:i+8], "%8b", &b)
		if err != nil {
			return "", err
		}
		builder.WriteByte(b)
	}
	return builder.String(), nil
}
