package auth

import "testing"

func TestRandomSecret(t *testing.T) {
	s1 := RandomSecret()
	s2 := RandomSecret()

	if s1 == "" || s2 == "" {
		t.Fatalf("expected non-empty secrets")
	}
	if s1 == s2 {
		t.Fatalf("expected secrets to differ")
	}
	// base64.RawURLEncoding should not include '=' padding.
	for _, s := range []string{s1, s2} {
		for i := 0; i < len(s); i++ {
			if s[i] == '=' {
				t.Fatalf("unexpected '=' in secret: %q", s)
			}
		}
	}
}
