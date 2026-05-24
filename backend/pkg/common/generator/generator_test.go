package generator

import (
	"strings"
	"testing"
)

func TestRandomStringIncludesRequestedLengthAndSymbols(t *testing.T) {
	got := RandomString(12)
	if len(got) != 12 {
		t.Fatalf("RandomString() length = %d, want 12", len(got))
	}

	if !strings.ContainsAny(got, symbolBytes) {
		t.Fatalf("RandomString() = %q, want at least one symbol", got)
	}
}

func TestRandomSimpleStringUsesRequestedLength(t *testing.T) {
	got := RandomSimpleString(16)
	if len(got) != 16 {
		t.Fatalf("RandomSimpleString() length = %d, want 16", len(got))
	}

	for _, r := range got {
		if !strings.ContainsRune(letterBytes, r) {
			t.Fatalf("RandomSimpleString() contains unexpected rune %q in %q", r, got)
		}
	}
}

func TestGenerateKeyReturnsMD5HexString(t *testing.T) {
	got, err := GenerateKey([]byte("lesson"))
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	if len(got) != 32 {
		t.Fatalf("GenerateKey() length = %d, want 32", len(got))
	}
}

func TestGenerateAPIKeyReturnsSHA256HexString(t *testing.T) {
	got, err := GenerateAPIKey("salt")
	if err != nil {
		t.Fatalf("GenerateAPIKey() error = %v", err)
	}
	if len(got) != 64 {
		t.Fatalf("GenerateAPIKey() length = %d, want 64", len(got))
	}
}

func TestGenerateNumbersInString(t *testing.T) {
	got := GenerateNumbersInString(6)
	if len(got) != 6 {
		t.Fatalf("GenerateNumbersInString() length = %d, want 6", len(got))
	}
	for _, r := range got {
		if r < '0' || r > '9' {
			t.Fatalf("GenerateNumbersInString() contains non-digit %q in %q", r, got)
		}
	}
}
