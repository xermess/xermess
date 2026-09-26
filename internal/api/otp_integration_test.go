package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// A login flow with the emailed code step holds a sign-in back: the password
// alone makes no session, the code in the message does, and a code that is
// wrong, spent or somebody else's does not.
func TestLiveEmailedCodeFinishesASignIn(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	// Six digits is the default; four makes the codes in this test shorter to
	// read, and proves the setting is the one the server actually uses.
	super.must(http.StatusOK, http.MethodPatch, "/otp", map[string]any{
		"length": 4, "max_attempts": 2, "resend_seconds": 0,
	}, nil)

	flow := defaultFlow(t, super)
	super.must(http.StatusOK, http.MethodPatch, "/login-flows/"+flow, map[string]any{
		"steps": []string{"identifier", "password", "email_code"},
	}, nil)

	super.must(http.StatusCreated, http.MethodPost, "/users", map[string]any{
		"email": "ada@example.com", "password": "ada-password-1", "confirm_password": "ada-password-1",
	}, nil)

	type challenge struct {
		Handle       string `json:"handle"`
		Email        string `json:"email"`
		AttemptsLeft int    `json:"attempts_left"`
	}
	type signedIn struct {
		Code *challenge `json:"code"`
	}

	b := s.browser()

	var held signedIn
	if status := b.account(http.MethodPost, "/login", map[string]string{
		"email": "ada@example.com", "password": "ada-password-1",
	}, &held); status != http.StatusOK {
		t.Fatalf("signing in = %d, want 200", status)
	}
	if held.Code == nil {
		t.Fatal("the password alone finished the sign-in; the flow asks for a code")
	}
	if held.Code.Handle == "" {
		t.Fatal("the sign-in was held with no handle to come back with")
	}

	// The address is shown back so somebody knows where to look, with the
	// middle of it hidden.
	if !strings.HasSuffix(held.Code.Email, "@example.com") || strings.Contains(held.Code.Email, "ada@") {
		t.Errorf("the address shown = %q, want it masked", held.Code.Email)
	}

	// Nobody is signed in until the code is typed.
	if status := b.account(http.MethodGet, "/me", nil, nil); status == http.StatusOK {
		t.Error("a session was made before the code was typed")
	}

	code := codeIn(t, s.mail.wait(t, "ada@example.com").Body, 4)

	// A wrong code is refused with guesses left, then the sign-in is over.
	var wrong problemBody
	if status := b.account(http.MethodPost, "/login/code", map[string]string{
		"handle": held.Code.Handle, "code": "0000",
	}, &wrong); status != http.StatusBadRequest || wrong.Code != "code_invalid" {
		t.Errorf("a wrong code = %d %+v, want code_invalid", status, wrong)
	}

	// The handle is the secret, so the right code under a handle nobody has
	// is worth nothing.
	var stolen problemBody
	if status := b.account(http.MethodPost, "/login/code", map[string]string{
		"handle": "not-a-handle", "code": code,
	}, &stolen); status != http.StatusGone || stolen.Code != "code_expired" {
		t.Errorf("a code under an unknown handle = %d %+v, want code_expired", status, stolen)
	}

	if status := b.account(http.MethodPost, "/login/code", map[string]string{
		"handle": held.Code.Handle, "code": code,
	}, nil); status != http.StatusOK {
		t.Fatalf("the code that was sent = %d, want 200", status)
	}

	if status := b.account(http.MethodGet, "/me", nil, nil); status != http.StatusOK {
		t.Error("the right code did not sign Ada in")
	}

	// The code is spent: the same one again finishes nothing.
	fresh := s.browser()
	var spent problemBody
	if status := fresh.account(http.MethodPost, "/login/code", map[string]string{
		"handle": held.Code.Handle, "code": code,
	}, &spent); status != http.StatusGone || spent.Code != "code_expired" {
		t.Errorf("a code used twice = %d %+v, want code_expired", status, spent)
	}
}

// A sign-in is over once the guesses run out, and the code it was waiting for
// stops working with it.
func TestLiveEmailedCodeRunsOutOfGuesses(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	super.must(http.StatusOK, http.MethodPatch, "/otp", map[string]any{
		"length": 4, "max_attempts": 1,
	}, nil)

	flow := defaultFlow(t, super)
	super.must(http.StatusOK, http.MethodPatch, "/login-flows/"+flow, map[string]any{
		"steps": []string{"identifier", "password", "email_code"},
	}, nil)

	super.must(http.StatusCreated, http.MethodPost, "/users", map[string]any{
		"email": "grace@example.com", "password": "grace-password-1", "confirm_password": "grace-password-1",
	}, nil)

	var held struct {
		Code *struct {
			Handle string `json:"handle"`
		} `json:"code"`
	}
	b := s.browser()
	b.account(http.MethodPost, "/login", map[string]string{
		"email": "grace@example.com", "password": "grace-password-1",
	}, &held)
	if held.Code == nil {
		t.Fatal("the sign-in was not held for a code")
	}

	code := codeIn(t, s.mail.wait(t, "grace@example.com").Body, 4)

	var used problemBody
	if status := b.account(http.MethodPost, "/login/code", map[string]string{
		"handle": held.Code.Handle, "code": "0000",
	}, &used); status != http.StatusGone || used.Code != "code_attempts_used" {
		t.Fatalf("the last guess = %d %+v, want code_attempts_used", status, used)
	}

	// Even the code that was sent: the sign-in it belonged to is over.
	var over problemBody
	if status := b.account(http.MethodPost, "/login/code", map[string]string{
		"handle": held.Code.Handle, "code": code,
	}, &over); status != http.StatusGone || over.Code != "code_expired" {
		t.Errorf("the right code after the guesses ran out = %d %+v, want code_expired", status, over)
	}
}

// Guesses cannot be multiplied by sending them at the same moment.
//
// Finding 17 in SECURITY-AUDIT-2.md: the count of wrong codes was read and
// written back, so requests arriving together all read the same number and
// every one of them got to compare a code. A six-digit code is a million
// guesses, and how many are allowed is the only thing standing in front of it.
func TestLiveEmailedCodeCountsGuessesSentAtOnce(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	const allowed = 3
	super.must(http.StatusOK, http.MethodPatch, "/otp", map[string]any{
		"length": 6, "max_attempts": allowed,
	}, nil)

	flow := defaultFlow(t, super)
	super.must(http.StatusOK, http.MethodPatch, "/login-flows/"+flow, map[string]any{
		"steps": []string{"identifier", "password", "email_code"},
	}, nil)

	super.must(http.StatusCreated, http.MethodPost, "/users", map[string]any{
		"email": "ada@example.com", "password": "ada-password-1", "confirm_password": "ada-password-1",
	}, nil)

	var held struct {
		Code *struct {
			Handle string `json:"handle"`
		} `json:"code"`
	}
	b := s.browser()
	b.account(http.MethodPost, "/login", map[string]string{
		"email": "ada@example.com", "password": "ada-password-1",
	}, &held)
	if held.Code == nil {
		t.Fatal("the sign-in was not held for a code")
	}

	sent := codeIn(t, s.mail.wait(t, "ada@example.com").Body, 6)

	// Twenty wrong codes at once, none of them the one that was sent.
	const guesses = 20
	wrong := make([]string, 0, guesses)
	for i := range guesses {
		guess := fmt.Sprintf("%06d", i)
		if guess == sent {
			guess = fmt.Sprintf("%06d", i+guesses)
		}
		wrong = append(wrong, guess)
	}

	answers := make(chan string, guesses)
	failures := make(chan error, guesses)
	release := make(chan struct{})

	var running sync.WaitGroup
	for _, guess := range wrong {
		running.Add(1)
		go func() {
			defer running.Done()

			<-release

			code, err := postCode(b, held.Code.Handle, guess)
			if err != nil {
				failures <- err
				return
			}
			answers <- code
		}()
	}

	close(release)
	running.Wait()
	close(answers)
	close(failures)

	for err := range failures {
		t.Fatalf("a guess never reached the server: %v", err)
	}

	counted := map[string]int{}
	for code := range answers {
		counted[code]++
	}

	// Every request that got as far as comparing a code and had guesses left
	// answers code_invalid; the one that spends the last guess, and everyone
	// refused before comparing, answers code_attempts_used. So the sign-in
	// allows `allowed` comparisons in all, whenever the requests arrive.
	if compared := counted["code_invalid"]; compared > allowed-1 {
		t.Errorf("%d of %d guesses sent at once were compared against the code, want at most %d: %v",
			compared, guesses, allowed-1, counted)
	}

	// And the sign-in is over, the code that was sent included.
	var over problemBody
	status := b.account(http.MethodPost, "/login/code", map[string]string{
		"handle": held.Code.Handle, "code": sent,
	}, &over)
	if status != http.StatusGone || (over.Code != "code_expired" && over.Code != "code_attempts_used") {
		t.Errorf("the right code after the guesses ran out = %d %+v, want the sign-in over", status, over)
	}
}

// postCode types one code back, from a goroutine of its own. It answers with
// the problem code and an error rather than calling t.Fatal, which only the
// goroutine running the test may do.
func postCode(b *browser, handle, code string) (string, error) {
	body, err := json.Marshal(map[string]string{"handle": handle, "code": code})
	if err != nil {
		return "", err
	}

	response, err := b.http.Post(b.s.root+"/api/v1/account/login/code", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var out problemBody
	if err := json.NewDecoder(response.Body).Decode(&out); err != nil {
		return "", err
	}

	return out.Code, nil
}

// Asking for another message moves the code, not the hour.
//
// Finding 20 in SECURITY-AUDIT-2.md: the deadline was written again on every
// resend, so a button pressed once a minute kept one sign-in alive for as long
// as somebody kept pressing it — while the comment above it said the opposite.
func TestLiveResendingACodeDoesNotExtendTheSignIn(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	// No wait between messages, so the second can be asked for at once.
	super.must(http.StatusOK, http.MethodPatch, "/otp", map[string]any{
		"length": 6, "resend_seconds": 0,
	}, nil)

	flow := defaultFlow(t, super)
	super.must(http.StatusOK, http.MethodPatch, "/login-flows/"+flow, map[string]any{
		"steps": []string{"identifier", "password", "email_code"},
	}, nil)

	super.must(http.StatusCreated, http.MethodPost, "/users", map[string]any{
		"email": "ada@example.com", "password": "ada-password-1", "confirm_password": "ada-password-1",
	}, nil)

	type challenge struct {
		Handle    string    `json:"handle"`
		ExpiresAt time.Time `json:"expires_at"`
	}

	var held struct {
		Code *challenge `json:"code"`
	}
	b := s.browser()
	b.account(http.MethodPost, "/login", map[string]string{
		"email": "ada@example.com", "password": "ada-password-1",
	}, &held)
	if held.Code == nil {
		t.Fatal("the sign-in was not held for a code")
	}

	first := codeIn(t, s.mail.wait(t, "ada@example.com").Body, 6)

	var again struct {
		Code challenge `json:"code"`
	}
	if status := b.account(http.MethodPost, "/login/code/resend", map[string]string{
		"handle": held.Code.Handle,
	}, &again); status != http.StatusOK {
		t.Fatalf("asking for another code = %d, want 200", status)
	}

	if !again.Code.ExpiresAt.Equal(held.Code.ExpiresAt) {
		t.Errorf("the sign-in expires at %s after a resend, was %s: the deadline moved",
			again.Code.ExpiresAt, held.Code.ExpiresAt)
	}

	second := codeIn(t, s.mail.nth(t, "ada@example.com", 2).Body, 6)

	// The message that came first is no longer the one being waited for. Two
	// codes drawn at random can be the same one, and then there is nothing to
	// say here.
	if second != first {
		var stale problemBody
		if status := b.account(http.MethodPost, "/login/code", map[string]string{
			"handle": held.Code.Handle, "code": first,
		}, &stale); status != http.StatusBadRequest || stale.Code != "code_invalid" {
			t.Errorf("the code from the first message = %d %+v, want code_invalid", status, stale)
		}
	}

	// The second one finishes the sign-in, inside the life the sign-in already
	// had.
	if status := b.account(http.MethodPost, "/login/code", map[string]string{
		"handle": held.Code.Handle, "code": second,
	}, nil); status != http.StatusOK {
		t.Fatalf("the code from the second message = %d, want 200", status)
	}
	if status := b.account(http.MethodGet, "/me", nil, nil); status != http.StatusOK {
		t.Error("the resent code did not sign Ada in")
	}
}

// codeIn reads the one-time code out of a message: the only run of exactly
// that many digits in it.
func codeIn(t *testing.T, body string, length int) string {
	t.Helper()

	matches := regexp.MustCompile(`\b\d{`+strconv.Itoa(length)+`}\b`).FindAllString(body, -1)
	if len(matches) != 1 {
		t.Fatalf("found %d codes in the message, want one:\n%s", len(matches), body)
	}

	return matches[0]
}
