package auth

import (
	"strings"
	"testing"
)

// The shape a code is read off paper in, and the shape check that tells one
// from an authenticator's six digits.
func TestNewRecoveryCodesAreTheShapeTheyAreReadIn(t *testing.T) {
	codes, hashes, err := newRecoveryCodes()
	if err != nil {
		t.Fatal(err)
	}

	if len(codes) != recoveryCodeCount || len(hashes) != recoveryCodeCount {
		t.Fatalf("got %d codes and %d hashes, want %d of each", len(codes), len(hashes), recoveryCodeCount)
	}

	seen := map[string]bool{}
	for i, code := range codes {
		half := recoveryCodeLetters / 2

		switch {
		case len(code) != recoveryCodeLetters+1:
			t.Errorf("code %q is %d characters, want %d", code, len(code), recoveryCodeLetters+1)
		case code[half] != '-':
			t.Errorf("code %q has no dash in the middle", code)
		case !isRecoveryCode(code):
			t.Errorf("code %q is not recognised as a recovery code", code)
		}

		for _, letter := range strings.ReplaceAll(code, "-", "") {
			if !strings.ContainsRune(recoveryAlphabet, letter) {
				t.Errorf("code %q has %q, which is not in the alphabet", code, letter)
			}
		}

		if seen[code] {
			t.Errorf("code %q was handed out twice", code)
		}
		seen[code] = true

		if hashes[i] != hashRecoveryCode(code) {
			t.Errorf("the hash stored for %q is not its own", code)
		}
	}
}

// Every letter has to be as likely as every other. Taking a byte modulo the
// alphabet made the first eight a ninth more common, which this is here to
// keep from coming back.
//
// It is a count rather than a proof, so the numbers are worth stating. Of the
// alphabet's 31 letters, 8 are the ones a modulo would favour: they should be
// 8/31 of what comes out, which is 0.258. The biased version produced 72/256,
// which is 0.281. Over this many letters the standard error is about 0.0014,
// so the gap of 0.023 is some sixteen of them — never a coincidence — while
// the bound below is nearly six, which an even spread will not cross in the
// life of this repository.
func TestRecoveryCodeLettersAreEvenlySpread(t *testing.T) {
	const (
		rounds   = 1000
		favoured = 256 % len(recoveryAlphabet) // 8: the letters a modulo would repeat
		expected = float64(favoured) / float64(len(recoveryAlphabet))
		bound    = 0.008
	)

	letters, hits := 0, 0

	for range rounds {
		codes, _, err := newRecoveryCodes()
		if err != nil {
			t.Fatal(err)
		}

		for _, code := range codes {
			for _, letter := range strings.ReplaceAll(code, "-", "") {
				letters++
				if strings.IndexRune(recoveryAlphabet, letter) < favoured {
					hits++
				}
			}
		}
	}

	share := float64(hits) / float64(letters)
	if share < expected-bound || share > expected+bound {
		t.Errorf("the first %d letters are %.4f of %d, want %.4f give or take %.3f",
			favoured, share, letters, expected, bound)
	}
}
