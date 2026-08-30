package logging

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	var gotID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = RequestIDFromContext(r.Context())
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	RequestID(next).ServeHTTP(rec, req)
	if gotID == "" {
		t.Fatal("request id not injected into context")
	}
	if rec.Header().Get("X-Request-ID") != gotID {
		t.Errorf("X-Request-ID = %q, want %q", rec.Header().Get("X-Request-ID"), gotID)
	}
}

func TestRequestIDHandler_AddsRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(&requestIDHandler{base: slog.NewJSONHandler(&buf, nil)})
	ctx := WithRequestID(context.Background(), "test-id-123")
	logger.InfoContext(ctx, "hello")
	out := buf.String()
	if !strings.Contains(out, `"request_id":"test-id-123"`) {
		t.Errorf("output missing request_id: %s", out)
	}
}

func TestRequestIDHandler_NoRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(&requestIDHandler{base: slog.NewJSONHandler(&buf, nil)})
	logger.InfoContext(context.Background(), "hello")
	if strings.Contains(buf.String(), "request_id") {
		t.Errorf("output should not contain request_id: %s", buf.String())
	}
}
