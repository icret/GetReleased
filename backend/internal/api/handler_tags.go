package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"getreleased/internal/database"
	"getreleased/internal/release"
)

func (a *API) handleListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := a.db.ListTags(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list tags: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tags)
}

func (a *API) handleCreateTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name required")
		return
	}
	tag, err := a.db.CreateTag(r.Context(), req.Name, req.Type)
	if err != nil {
		if errors.Is(err, database.ErrDuplicate) {
			writeError(w, http.StatusConflict, "tag already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "create tag: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tag)
}

func (a *API) handleUpdateTag(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name required")
		return
	}
	if err := a.db.UpdateTag(r.Context(), id, req.Name, req.Type); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "tag not found")
			return
		}
		if database.IsUniqueConstraintError(err) {
			writeError(w, http.StatusConflict, "tag name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "update tag: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, release.Tag{
		ID:   id,
		Name: req.Name,
		Type: database.NormalizeTagType(req.Type),
	})
}

func (a *API) handleDeleteTag(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := a.db.DeleteTag(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "tag not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "delete tag: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}
