package id

import (
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name        string
		codeLength  int
		expectedLen int
	}{
		{"default length", 0, 8},
		{"custom length", 12, 12},
		{"negative length", -5, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(tt.codeLength)
			if gen.codeLength != tt.expectedLen {
				t.Errorf("expected code length %d, got %d", tt.expectedLen, gen.codeLength)
			}
		})
	}
}

func TestGenerateCode(t *testing.T) {
	gen := NewGenerator(8)
	
	// Generate multiple codes and check uniqueness
	codes := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		code := gen.GenerateCode()
		
		// Check length
		if len(code) != 8 {
			t.Errorf("expected code length 8, got %d", len(code))
		}
		
		// Check if code contains only valid characters
		for _, char := range code {
			if !strings.ContainsRune(base62Chars, char) {
				t.Errorf("code contains invalid character: %c", char)
			}
		}
		
		// Check uniqueness
		if codes[code] {
			t.Errorf("duplicate code generated: %s", code)
		}
		codes[code] = true
	}
}

func TestGenerateCustomCode(t *testing.T) {
	gen := NewGenerator(8)
	
	tests := []struct {
		name     string
		custom   string
		expected string
	}{
		{"empty string", "", ""}, // Will generate random code
		{"valid custom", "myurl", "myurl"},
		{"with spaces", " my url ", "myurl"},
		{"with special chars", "my@url!", "myurl"},
		{"too short", "abc", "abc"},
		{"too long", "verylongurlcode", "verylongu"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := gen.GenerateCustomCode(tt.custom)
			
			if tt.custom == "" {
				// Should generate a random code
				if len(code) != 8 {
					t.Errorf("expected random code length 8, got %d", len(code))
				}
			} else {
				// Should use custom code (cleaned)
				expected := strings.TrimSpace(tt.custom)
				if len(expected) < 8 {
					expected = expected + gen.generateRandomSuffix(8-len(expected))
				}
				if len(expected) > 8 {
					expected = expected[:8]
				}
				
				if code != expected {
					t.Errorf("expected %s, got %s", expected, code)
				}
			}
		})
	}
}

func TestValidateCode(t *testing.T) {
	gen := NewGenerator(8)
	
	tests := []struct {
		name  string
		code  string
		valid bool
	}{
		{"valid code", "abc123", true},
		{"valid code with mixed case", "AbC123", true},
		{"too short", "abc", false},
		{"too long", "abcdefghijklmnop", false},
		{"invalid characters", "abc@123", false},
		{"empty string", "", false},
		{"with spaces", "abc 123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.ValidateCode(tt.code)
			if result != tt.valid {
				t.Errorf("expected validation result %v for code '%s', got %v", tt.valid, tt.code, result)
			}
		})
	}
}

func TestToBase62(t *testing.T) {
	gen := NewGenerator(8)
	
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"zero bytes", []byte{}, ""},
		{"single byte", []byte{0}, "0"},
		{"multiple bytes", []byte{1, 2, 3}, "321"},
		{"large number", []byte{255, 255, 255}, "777777"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.toBase62(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGenerateRandomSuffix(t *testing.T) {
	gen := NewGenerator(8)
	
	// Test different lengths
	for length := 1; length <= 10; length++ {
		suffix := gen.generateRandomSuffix(length)
		
		if len(suffix) != length {
			t.Errorf("expected suffix length %d, got %d", length, len(suffix))
		}
		
		// Check if all characters are valid
		for _, char := range suffix {
			if !strings.ContainsRune(base62Chars, char) {
				t.Errorf("suffix contains invalid character: %c", char)
			}
		}
	}
	
	// Test zero length
	suffix := gen.generateRandomSuffix(0)
	if suffix != "" {
		t.Errorf("expected empty string for zero length, got %s", suffix)
	}
}

func TestSanitizeChar(t *testing.T) {
	gen := NewGenerator(8)
	
	tests := []struct {
		name     string
		input    rune
		expected bool
	}{
		{"digit", '5', true},
		{"uppercase letter", 'A', true},
		{"lowercase letter", 'z', true},
		{"special character", '@', false},
		{"space", ' ', false},
		{"newline", '\n', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.sanitizeChar(tt.input)
			
			if tt.expected {
				// Should return the same character
				if result != tt.input {
					t.Errorf("expected %c, got %c", tt.input, result)
				}
			} else {
				// Should return a valid character
				if !strings.ContainsRune(base62Chars, result) {
					t.Errorf("result %c is not a valid base62 character", result)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkGenerateCode(b *testing.B) {
	gen := NewGenerator(8)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateCode()
	}
}

func BenchmarkGenerateCustomCode(b *testing.B) {
	gen := NewGenerator(8)
	custom := "myurl"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateCustomCode(custom)
	}
}

func BenchmarkValidateCode(b *testing.B) {
	gen := NewGenerator(8)
	code := "abc123"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.ValidateCode(code)
	}
}
