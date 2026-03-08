package random

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
	"testing"
)

// randomSuffix generates a random lowercase string suffix for creating unique test identifiers.
//
// Implementation choices:
//   - base32 hex encoding: Creates DNS/hostname-safe identifiers (only 0-9, a-v)
//   - Lowercase: Ensures compatibility with case-insensitive systems (Windows, some DBs)
//     and satisfies cloud provider naming requirements (GCP, AWS, Kubernetes)
//   - No padding: Removes trailing '=' characters for cleaner identifiers and avoids
//     issues where padding might be URL-encoded or misinterpreted
//   - 10 bytes input: Generates ~16 character output, balancing uniqueness with brevity
//
// The result is safe for use in:
//   - DNS names and hostnames
//   - Cloud resource names (Google Cloud Spanner, Cloud SQL, AWS resources, etc.)
//   - Kubernetes resource identifiers
//   - URLs and file paths (no escaping needed)
func randomSuffix(t testing.TB) string {
	data := make([]byte, 10)
	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}
	return strings.ToLower(base32.HexEncoding.WithPadding(base32.NoPadding).EncodeToString(data))
}

func TestRandomSuffix(t *testing.T) {
	suffix := randomSuffix(t)
	if len(suffix) != 16 {
		t.Errorf("expected suffix length of 16, got %d", len(suffix))
	}
	for _, r := range suffix {
		if (r < 'a' || r > 'v') && (r < '0' || r > '9') {
			t.Errorf("unexpected character %q in suffix", r)
		}
	}
}
