package api

import (
	"context"
	"net/http"
	"time"

	"getreleased/internal/exporter"

	"github.com/google/uuid"
)

type trackState struct {
	running    bool
	lastTaskID string
	startedAt  time.Time
	finishedAt time.Time
	dirty      bool
	err        string
}

type trackStatusResponse struct {
	Running    bool       `json:"running"`
	LastTaskID string     `json:"last_task_id"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Dirty      bool       `json:"dirty,omitempty"`
	Error      string     `json:"error,omitempty"`
}

func (a *API) handleTrack(w http.ResponseWriter, r *http.Request) {
	a.trackMu.Lock()
	if a.trackState.running {
		a.trackMu.Unlock()
		writeError(w, http.StatusConflict, "track already running")
		return
	}
	taskID := uuid.NewString()
	a.trackState = trackState{
		running:    true,
		lastTaskID: taskID,
		startedAt:  time.Now(),
	}
	a.trackMu.Unlock()

	go func() {
		ctx := context.WithoutCancel(r.Context())
		dirty, err := a.trk.Track(ctx)
		a.trackMu.Lock()
		a.trackState.running = false
		a.trackState.finishedAt = time.Now()
		a.trackState.dirty = dirty
		a.trackState.err = ""
		if err != nil {
			a.trackState.err = err.Error()
		}
		a.trackMu.Unlock()
	}()

	writeJSON(w, http.StatusOK, map[string]string{"task_id": taskID})
}

func (a *API) handleExport(w http.ResponseWriter, r *http.Request) {
	a.exportMu.Lock()
	defer a.exportMu.Unlock()
	if err := exporter.Export(r.Context(), a.db, a.exportDir); err != nil {
		writeError(w, http.StatusInternalServerError, "export: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"exported": true})
}

func (a *API) handleTrackStatus(w http.ResponseWriter, r *http.Request) {
	a.trackMu.Lock()
	state := a.trackState
	a.trackMu.Unlock()

	resp := trackStatusResponse{
		Running:    state.running,
		LastTaskID: state.lastTaskID,
		Dirty:      state.dirty,
		Error:      state.err,
	}
	if !state.startedAt.IsZero() {
		resp.StartedAt = &state.startedAt
	}
	if !state.finishedAt.IsZero() {
		resp.FinishedAt = &state.finishedAt
	}
	writeJSON(w, http.StatusOK, resp)
}
