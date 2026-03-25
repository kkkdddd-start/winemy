//go:build windows

package codec

import (
	"testing"
)

func TestEncodeBase64(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "Hello", "SGVsbG8="},
		{"world", "World", "V29ybGQ="},
		{"empty", "", ""},
		{"numbers", "12345", "MTIzNDU="},
		{"special", "!@#$%", "IUAjJCU="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Encode(tt.input, CodecBase64)
			if err != nil {
				t.Errorf("Encode() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Encode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDecodeBase64(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"simple", "SGVsbG8=", "Hello", false},
		{"world", "V29ybGQ=", "World", false},
		{"invalid", "not-valid-base64!!!", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Decode(tt.input, CodecBase64)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("Decode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeHex(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"hello", "Hello", "48656c6c6f"},
		{"abc", "ABC", "414243"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Encode(tt.input, CodecHex)
			if err != nil {
				t.Errorf("Encode() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Encode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDecodeHex(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"hello", "48656c6c6f", "Hello", false},
		{"upper", "414243", "ABC", false},
		{"invalid", "not-hex", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Decode(tt.input, CodecHex)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("Decode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeURL(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "hello world", "hello+world"},
		{"special", "a=b&c=d", "a%3Db%26c%3Dd"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Encode(tt.input, CodecURL)
			if err != nil {
				t.Errorf("Encode() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Encode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeHTML(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"html", "<div>", "&lt;div&gt;"},
		{"ampersand", "a&b", "a&amp;b"},
		{"quotes", "\"test\"", "&quot;test&quot;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Encode(tt.input, CodecHTML)
			if err != nil {
				t.Errorf("Encode() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Encode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeBinary(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"a", "A", "01000001"},
		{"hello", "Hello", "01001000 01100101 01101100 01101100 01101111"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Encode(tt.input, CodecBinary)
			if err != nil {
				t.Errorf("Encode() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Encode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAutoDetect(t *testing.T) {
	engine := New()

	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{"base64", "SGVsbG8=", 1},
		{"hex", "48656c6c6f", 1},
		{"unicode", "\\u0048\\u0065", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := engine.AutoDetect(tt.input)
			if err != nil {
				t.Errorf("AutoDetect() error = %v", err)
				return
			}
			if len(results) < tt.wantLen {
				t.Errorf("AutoDetect() returned %d results, want at least %d", len(results), tt.wantLen)
			}
		})
	}
}

func TestCodecRoundTrip(t *testing.T) {
	engine := New()

	types := []CodecType{CodecBase64, CodecHex, CodecURL, CodecHTML, CodecBinary}

	for _, ct := range types {
		t.Run(string(ct), func(t *testing.T) {
			encoded, err := engine.Encode("test data 123", ct)
			if err != nil {
				t.Errorf("Encode(%s) error = %v", ct, err)
				return
			}

			decoded, err := engine.Decode(encoded, ct)
			if err != nil {
				t.Errorf("Decode(%s) error = %v", ct, err)
				return
			}

			if decoded != "test data 123" {
				t.Errorf("Decode(Encode(%s)) = %v, want test data 123", ct, decoded)
			}
		})
	}
}

func TestDecodeUnicode(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"simple", "\\u0048\\u0065", "He", false},
		{"hello", "\\u0048\\u0065\\u006c\\u006c\\u006f", "Hello", false},
		{"invalid", "not-unicode", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Decode(tt.input, CodecUnicode)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("Decode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDecodeURL(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"simple", "hello+world", "hello world", false},
		{"encoded", "a%3Db%26c%3Dd", "a=b&c=d", false},
		{"invalid", "%ZZ", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Decode(tt.input, CodecURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("Decode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDecodeBinary(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"a", "01000001", "A", false},
		{"hello", "01001000 01100101 01101100 01101100 01101111", "Hello", false},
		{"invalid length", "0100000", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Decode(tt.input, CodecBinary)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("Decode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeBase64URL(t *testing.T) {
	engine := New()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "Hello", "SGVsbG8="},
		{"url unsafe", "a+b/c", "YStiYy9j"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Encode(tt.input, CodecBase64URL)
			if err != nil {
				t.Errorf("Encode() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Encode() = %v, want %v", result, tt.expected)
			}
		})
	}
}
