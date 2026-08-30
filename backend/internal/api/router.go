package api

import (
	"net/http"
	"sync"

	"getreleased/internal/auth"
	"getreleased/internal/database"
	"getreleased/internal/github"
	"getreleased/internal/logging"
	"getreleased/internal/tracker"
)

type API struct {
	db         *database.DB
	trk        *tracker.Tracker
	ghClient   github.Fetcher
	authSvc    *auth.Service
	exportDir  string
	exportMu   sync.Mutex
	trackMu    sync.Mutex
	trackState trackState
}

func New(db *database.DB, trk *tracker.Tracker, ghClient github.Fetcher, authSvc *auth.Service, exportDir string) *API {
	return &API{db: db, trk: trk, ghClient: ghClient, authSvc: authSvc, exportDir: exportDir}
}

func (a *API) Router(isDev bool) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/login", a.handleLogin)

	admin := http.NewServeMux()
	admin.HandleFunc("GET /api/admin/repositories", a.handleListRepositories)
	admin.HandleFunc("POST /api/admin/repositories", a.handleCreateRepository)
	admin.HandleFunc("PUT /api/admin/repositories/{id}", a.handleUpdateRepository)
	admin.HandleFunc("DELETE /api/admin/repositories/{id}", a.handleDeleteRepository)
	admin.HandleFunc("PUT /api/admin/repositories/{id}/tags", a.handleSetRepositoryTags)
	admin.HandleFunc("POST /api/admin/repositories/{id}/sync", a.handleSyncRepository)
	admin.HandleFunc("GET /api/admin/tags", a.handleListTags)
	admin.HandleFunc("POST /api/admin/tags", a.handleCreateTag)
	admin.HandleFunc("PUT /api/admin/tags/{id}", a.handleUpdateTag)
	admin.HandleFunc("DELETE /api/admin/tags/{id}", a.handleDeleteTag)
	admin.HandleFunc("POST /api/admin/track", a.handleTrack)
	admin.HandleFunc("POST /api/admin/export", a.handleExport)
	admin.HandleFunc("GET /api/admin/track/status", a.handleTrackStatus)
	admin.HandleFunc("GET /api/admin/stats", a.handleStats)
	admin.HandleFunc("GET /api/admin/users", a.handleListUsers)
	admin.HandleFunc("POST /api/admin/users", a.handleCreateUser)
	admin.HandleFunc("PUT /api/admin/users/{id}/password", a.handleResetPassword)
	admin.HandleFunc("DELETE /api/admin/users/{id}", a.handleDeleteUser)

	mux.Handle("/api/admin/", auth.RequireAuth(a.authSvc.JWTSecret())(admin))

	h := http.Handler(mux)
	h = logging.RequestID(h)
	if isDev {
		h = withCORS(h)
	}
	return h
}
