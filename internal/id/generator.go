package id

import (
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/rs/xid"
)

const (
	// Base62 characters for URL-friendly encoding
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// Default code length for short URLs
	defaultCodeLength = 8
)

// Generator provides ID generation functionality
type Generator struct {
	codeLength int
}

// NewGenerator creates a new ID generator
func NewGenerator(codeLength int) *Generator {
	if codeLength <= 0 {
		codeLength = defaultCodeLength
	}
	return &Generator{codeLength: codeLength}
}

// GenerateCode generates a unique short code for URLs
func (g *Generator) GenerateCode() string {
	// Use ULID for uniqueness and time ordering
	id := xid.New()
	
	// Convert to base62 for URL-friendly encoding
	code := g.toBase62(id.Bytes())
	
	// Ensure minimum length
	if len(code) < g.codeLength {
		code = code + g.generateRandomSuffix(g.codeLength-len(code))
	}
	
	// Truncate to desired length
	if len(code) > g.codeLength {
		code = code[:g.codeLength]
	}
	
	return code
}

// GenerateCustomCode generates a custom code with validation
func (g *Generator) GenerateCustomCode(custom string) string {
	// Clean and validate custom code
	clean := strings.TrimSpace(custom)
	if len(clean) == 0 {
		return g.GenerateCode()
	}
	
	// Ensure it's URL-safe
	clean = strings.Map(g.sanitizeChar, clean)
	
	// Ensure minimum length
	if len(clean) < g.codeLength {
		clean = clean + g.generateRandomSuffix(g.codeLength-len(clean))
	}
	
	// Truncate to desired length
	if len(clean) > g.codeLength {
		clean = clean[:g.codeLength]
	}
	
	return clean
}

// toBase62 converts bytes to base62 string
func (g *Generator) toBase62(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	
	// Convert to big.Int for base conversion
	var num big.Int
	num.SetBytes(data)
	
	// Convert to base62
	base := big.NewInt(62)
	var result strings.Builder
	
	for num.Sign() > 0 {
		remainder := new(big.Int)
		num.DivMod(&num, base, remainder)
		result.WriteByte(base62Chars[remainder.Int64()])
	}
	
	// Reverse the result
	runes := []rune(result.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	
	return string(runes)
}

// generateRandomSuffix generates a random suffix of specified length
func (g *Generator) generateRandomSuffix(length int) string {
	if length <= 0 {
		return ""
	}
	
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		// Generate random index in base62 range
		idx, err := rand.Int(rand.Reader, big.NewInt(62))
		if err != nil {
			// Fallback to timestamp-based generation
			idx = big.NewInt(time.Now().UnixNano() % 62)
		}
		result[i] = base62Chars[idx.Int64()]
	}
	
	return string(result)
}

// sanitizeChar ensures characters are URL-safe
func (g *Generator) sanitizeChar(r rune) rune {
	if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
		return r
	}
	// Replace invalid characters with random base62 char
	return rune(base62Chars[time.Now().UnixNano()%62])
}

// ValidateCode validates if a code meets requirements
func (g *Generator) ValidateCode(code string) bool {
	if len(code) < 4 || len(code) > 16 {
		return false
	}
	
	// Check if all characters are valid
	for _, r := range code {
		if !strings.ContainsRune(base62Chars, r) {
			return false
		}
	}
	
	return true
}
