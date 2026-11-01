package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Push-to-deploy, the same shape as the other pflow.dev apps: GitHub POSTs
// /webhook/github, the running binary verifies the HMAC, fetches, refuses an
// unsigned origin/main, fast-forwards, rebuilds, then restarts itself through
// ~/services and exits. It answers 503 while DEPLOY_SECRET is unset.

const serviceName = "mine-pflow-xyz"

var deployMu sync.Mutex

// verifyHMAC checks GitHub's X-Hub-Signature-256 header against the body.
func verifyHMAC(secret string, body []byte, signature string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	sig, err := hex.DecodeString(signature[len("sha256="):])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(sig, mac.Sum(nil))
}

func (a *App) handleWebhook(w http.ResponseWriter, r *http.Request) {
	secret := os.Getenv("DEPLOY_SECRET")
	if secret == "" {
		http.Error(w, "deploy not configured", http.StatusServiceUnavailable)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if !verifyHMAC(secret, body, r.Header.Get("X-Hub-Signature-256")) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}
	if r.Header.Get("X-GitHub-Event") == "ping" {
		fmt.Fprint(w, "pong")
		return
	}
	var payload struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if payload.Ref != "refs/heads/main" {
		fmt.Fprintf(w, "ignored: %s", payload.Ref)
		return
	}
	go func() {
		if !deployMu.TryLock() {
			log.Printf("[deploy] already running, skipping")
			return
		}
		defer deployMu.Unlock()
		log.Printf("[deploy] webhook triggered for %s", payload.Ref)
		out, err := a.deploy()
		if err != nil {
			log.Printf("[deploy] failed: %v\n%s", err, out)
			return
		}
		log.Printf("[deploy] success, scheduling restart")
		scheduleRestart()
	}()
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "deploy started")
}

// runDeploy fetches, verifies, fast-forwards and rebuilds the checkout the
// service runs from (its working directory).
func runDeploy() (string, error) {
	var buf strings.Builder
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	run := func(desc string, args ...string) error {
		fmt.Fprintf(&buf, "==> %s\n", desc)
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		buf.Write(out)
		if err != nil {
			return fmt.Errorf("%s: %w", desc, err)
		}
		return nil
	}
	if err := run("git fetch", "git", "fetch", "origin", "main"); err != nil {
		return buf.String(), err
	}
	if err := run("verify signature", "git", "verify-commit", "origin/main"); err != nil {
		return buf.String(), fmt.Errorf("refusing to deploy unsigned/untrusted HEAD: %w", err)
	}
	if err := run("fast-forward", "git", "merge", "--ff-only", "origin/main"); err != nil {
		return buf.String(), err
	}
	if err := run("build", "make", "build"); err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}

// scheduleRestart hands the restart to ~/services and exits so the new binary
// takes over the tmux window.
func scheduleRestart() {
	time.Sleep(2 * time.Second)
	home, _ := os.UserHomeDir()
	cmd := exec.Command("nohup", filepath.Join(home, "services"), "restart", serviceName)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("[deploy] restart failed: %v", err)
		return
	}
	_ = cmd.Process.Release()
	log.Printf("[deploy] exiting for restart")
	os.Exit(0)
}
