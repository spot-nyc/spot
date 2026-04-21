package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRunUpdateCheck_PrintsMessageWhenNewer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"tag_name":"v0.99.0"}`)
	}))
	defer srv.Close()

	var stderr bytes.Buffer
	cacheDir := t.TempDir()

	runUpdateCheckWithOptions(context.Background(), updateCheckOptions{
		currentVersion: "v0.1.0",
		cacheDir:       cacheDir,
		apiBaseURL:     srv.URL,
		grace:          2 * time.Second,
		stderr:         &stderr,
	})

	got := stderr.String()
	if got == "" {
		t.Fatal("expected update message on stderr, got empty")
	}
	if !bytes.Contains(stderr.Bytes(), []byte("v0.99.0")) {
		t.Errorf("stderr = %q, want it to mention v0.99.0", got)
	}
}

func TestRunUpdateCheck_SilentWhenSameVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"tag_name":"v0.1.0"}`)
	}))
	defer srv.Close()

	var stderr bytes.Buffer
	runUpdateCheckWithOptions(context.Background(), updateCheckOptions{
		currentVersion: "v0.1.0",
		cacheDir:       t.TempDir(),
		apiBaseURL:     srv.URL,
		grace:          2 * time.Second,
		stderr:         &stderr,
	})

	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty, got %q", stderr.String())
	}
}

func TestRunUpdateCheck_GraceTimeoutDoesNotPrint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = io.WriteString(w, `{"tag_name":"v0.99.0"}`)
	}))
	defer srv.Close()

	var stderr bytes.Buffer
	runUpdateCheckWithOptions(context.Background(), updateCheckOptions{
		currentVersion: "v0.1.0",
		cacheDir:       t.TempDir(),
		apiBaseURL:     srv.URL,
		grace:          10 * time.Millisecond,
		stderr:         &stderr,
	})

	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty when grace expires, got %q", stderr.String())
	}
}
