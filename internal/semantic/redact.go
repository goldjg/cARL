package semantic

import (
	"regexp"
	"strings"
)

var secretLinePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|token|secret|password|passwd|credential|authorization)\s*[:=]`),
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._~+/=-]{12,}`),
	regexp.MustCompile(`(?i)-----BEGIN [A-Z ]*PRIVATE KEY-----`),
	regexp.MustCompile(`(?i)github_pat_[A-Za-z0-9_]+`),
	regexp.MustCompile(`(?i)gh[pousr]_[A-Za-z0-9_]{20,}`),
	regexp.MustCompile(`(?i)xox[baprs]-[A-Za-z0-9-]{20,}`),
}

var secretPathPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(^|/)\.env(\.|$)`),
	regexp.MustCompile(`(?i)(^|/)(id_rsa|id_dsa|id_ecdsa|id_ed25519)(\.|$)`),
	regexp.MustCompile(`(?i)(secret|credential|private[-_]?key|token|password)`),
	regexp.MustCompile(`(?i)\.(pem|p12|pfx|key)$`),
}

// RedactText redacts obvious secret-bearing lines and reports whether any
// content changed.
func RedactText(input string) (string, bool) {
	var out []string
	redacted := false
	for _, line := range strings.Split(input, "\n") {
		if isSecretLine(line) {
			out = append(out, "[REDACTED]")
			redacted = true
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n"), redacted
}

// IsSecretPath reports whether a repository path is likely to contain secret
// material and should not have contents sent to external evaluators.
func IsSecretPath(path string) bool {
	normalized := strings.ReplaceAll(path, "\\", "/")
	for _, re := range secretPathPatterns {
		if re.MatchString(normalized) {
			return true
		}
	}
	return false
}

func isSecretLine(line string) bool {
	for _, re := range secretLinePatterns {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}
