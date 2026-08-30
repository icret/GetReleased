package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"getreleased/internal/database"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	svc, err := NewService(db, "jwtkey", 12*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func seedAdmin(t *testing.T, svc *Service, username, password string) {
	t.Helper()
	if err := svc.EnsureAdminSeed(context.Background(), username, password); err != nil {
		t.Fatal(err)
	}
}

func TestNewServiceValidation(t *testing.T) {
	dir := t.TempDir()
	db, err := database.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := db.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		jwtSecret string
		wantErr   bool
	}{
		{name: "empty_secret", wantErr: true},
		{name: "ok", jwtSecret: "s", wantErr: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewService(db, tc.jwtSecret, 12*time.Hour)
			if (err != nil) != tc.wantErr {
				t.Errorf("expected err=%v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestEnsureAdminSeedIdempotent(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	seedAdmin(t, svc, "admin", "secret123")
	count, err := svc.db.CountUsers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 user, got %d", count)
	}

	if err := svc.EnsureAdminSeed(ctx, "admin", "other"); err != nil {
		t.Fatal(err)
	}
	count, _ = svc.db.CountUsers(ctx)
	if count != 1 {
		t.Errorf("expected 1 user after re-seed, got %d", count)
	}
}

func TestEnsureAdminSeedEmptyPassword(t *testing.T) {
	svc := newTestService(t)
	err := svc.EnsureAdminSeed(context.Background(), "admin", "")
	if err == nil {
		t.Error("expected error for empty password with no users")
	}
}

func TestVerifyPassword(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	seedAdmin(t, svc, "admin", "secret123")

	if !svc.VerifyPassword(ctx, "admin", "secret123") {
		t.Error("expected valid password")
	}
	if svc.VerifyPassword(ctx, "admin", "wrong") {
		t.Error("expected invalid password")
	}
	if svc.VerifyPassword(ctx, "nobody", "secret123") {
		t.Error("expected invalid for missing user")
	}
}

func TestIssueAndParseToken(t *testing.T) {
	svc := newTestService(t)
	token, exp, err := svc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Error("empty token")
	}
	subject, err := ParseToken(token, []byte("jwtkey"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if subject != "admin" {
		t.Errorf("expected admin, got %s", subject)
	}
	if !exp.After(time.Now()) {
		t.Error("expiry in past")
	}
}

func TestParseTokenExpired(t *testing.T) {
	dir := t.TempDir()
	db, err := database.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := db.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	svc, err := NewService(db, "jwtkey", -1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	token, _, _ := svc.IssueToken("admin")
	if _, err := ParseToken(token, []byte("jwtkey")); err == nil {
		t.Error("expected error for expired token")
	}
}

func TestParseTokenWrongSecret(t *testing.T) {
	svc := newTestService(t)
	token, _, _ := svc.IssueToken("admin")
	if _, err := ParseToken(token, []byte("wrong")); err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestRequireAuth(t *testing.T) {
	svc := newTestService(t)
	token, _, _ := svc.IssueToken("admin")

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{name: "no_header", authHeader: "", wantStatus: http.StatusUnauthorized},
		{name: "bad_scheme", authHeader: "Basic abc", wantStatus: http.StatusUnauthorized},
		{name: "bad_token", authHeader: "Bearer invalid", wantStatus: http.StatusUnauthorized},
		{name: "ok", authHeader: "Bearer " + token, wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := RequireAuth(svc.JWTSecret())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Errorf("expected %d, got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "" {
		t.Fatal("empty hash")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte("secret123")) != nil {
		t.Error("hash does not match password")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte("wrong")) == nil {
		t.Error("hash should not match wrong password")
	}
}
