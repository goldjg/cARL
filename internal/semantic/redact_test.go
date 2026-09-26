package semantic

import (
	"strings"
	"testing"
)

func TestRedactTextRemovesSecretLines(t *testing.T) {
	out, redacted := RedactText("safe\napi_key = super-secret\nAuthorization: Bearer abcdefghijklmnop\n")
	if !redacted {
		t.Fatal("expected redaction")
	}
	if strings.Contains(out, "super-secret") || strings.Contains(out, "abcdefghijklmnop") {
		t.Fatalf("secret leaked in redacted output: %q", out)
	}
}

func TestIsSecretPath(t *testing.T) {
	for _, path := range []string{".env", "config/private-key.pem", "secrets/token.txt"} {
		if !IsSecretPath(path) {
			t.Fatalf("expected secret path: %s", path)
		}
	}
	if IsSecretPath("internal/review/review.go") {
		t.Fatal("ordinary source file should not be secret-bearing")
	}
}
