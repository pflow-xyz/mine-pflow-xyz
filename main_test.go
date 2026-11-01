package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	(&App{}).routes().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestHealth(t *testing.T) {
	if w := get(t, "/health"); w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Fatalf("/health = %d %q", w.Code, w.Body.String())
	}
}

func TestLandingServedAndUnknownPathsAre404(t *testing.T) {
	w := get(t, "/")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "<title>") {
		t.Fatalf("/ = %d", w.Code)
	}
	if w := get(t, "/nope"); w.Code != http.StatusNotFound {
		t.Fatalf("/nope = %d, want 404", w.Code)
	}
}

// The public copy leads with outcomes; the engine is not the pitch. README.md
// is where the engine gets its one sentence.
func TestLandingCopyDoesNotPitchTheEngine(t *testing.T) {
	err := fs.WalkDir(webFS, "web", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := webFS.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(strings.ToLower(string(b)), "petri") {
			t.Errorf("%s mentions the engine by name", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func sign(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func postWebhook(a *App, body, signature, event string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/webhook/github", strings.NewReader(body))
	r.Header.Set("X-Hub-Signature-256", signature)
	r.Header.Set("X-GitHub-Event", event)
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	return w
}

func TestWebhookWithoutSecretIsUnavailable(t *testing.T) {
	t.Setenv("DEPLOY_SECRET", "")
	if w := postWebhook(&App{}, "{}", "", "push"); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d", w.Code)
	}
}

func TestWebhookRejectsBadSignature(t *testing.T) {
	t.Setenv("DEPLOY_SECRET", "s3")
	called := make(chan struct{}, 1)
	a := &App{deploy: func() (string, error) { called <- struct{}{}; return "", nil }}
	body := `{"ref":"refs/heads/main"}`
	if w := postWebhook(a, body, sign("wrong", body), "push"); w.Code != http.StatusForbidden {
		t.Fatalf("got %d", w.Code)
	}
	select {
	case <-called:
		t.Fatal("deploy ran on a bad signature")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestWebhookPingAndOtherBranchesDoNotDeploy(t *testing.T) {
	t.Setenv("DEPLOY_SECRET", "s3")
	called := make(chan struct{}, 2)
	a := &App{deploy: func() (string, error) { called <- struct{}{}; return "", nil }}

	if w := postWebhook(a, "{}", sign("s3", "{}"), "ping"); w.Code != http.StatusOK || w.Body.String() != "pong" {
		t.Fatalf("ping = %d %q", w.Code, w.Body.String())
	}
	other := `{"ref":"refs/heads/feature"}`
	if w := postWebhook(a, other, sign("s3", other), "push"); w.Code != http.StatusOK {
		t.Fatalf("other branch = %d", w.Code)
	}
	select {
	case <-called:
		t.Fatal("deploy ran for a ping or a non-main branch")
	case <-time.After(50 * time.Millisecond):
	}
}
