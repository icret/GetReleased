package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"getreleased/internal/exporter"

	gh "github.com/google/go-github/v69/github"
)

func isGitHubNotFound(err error) bool {
	var ghErr *gh.ErrorResponse
	if errors.As(err, &ghErr) {
		return ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound
	}
	return false
}

func (a *API) handleListRepositories(w http.ResponseWriter, r *http.Request) {
	repos, err := a.db.ListRepositories(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, repos)
}

func (a *API) handleCreateRepository(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Owner string `json:"owner"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Owner == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "owner and name required")
		return
	}

	result, err := a.ghClient.FetchRepository(r.Context(), req.Owner, req.Name, "", "")
	if err != nil {
		if isGitHubNotFound(err) {
			writeError(w, http.StatusBadRequest, "repository not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "fetch repository: "+err.Error())
		return
	}
	repoInfo := result.Repo

	fullName := repoInfo.GetFullName()
	existing, err := a.db.GetRepositoryByFullName(r.Context(), fullName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "check duplicate: "+err.Error())
		return
	}
	if existing != nil {
		writeError(w, http.StatusConflict, "repository already tracked")
		return
	}

	actualOwner := repoInfo.GetOwner().GetLogin()
	actualName := repoInfo.GetName()
	if _, err := a.trk.TrackOne(r.Context(), actualOwner, actualName); err != nil {
		writeError(w, http.StatusInternalServerError, "track: "+err.Error())
		return
	}
	if err := exporter.Export(r.Context(), a.db, a.exportDir); err != nil {
		writeError(w, http.StatusInternalServerError, "export: "+err.Error())
		return
	}

	repo, err := a.db.GetRepositoryByFullName(r.Context(), fullName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "fetch created: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, repo)
}

func (a *API) handleUpdateRepository(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req struct {
		Description string `json:"description"`
		Remark      string `json:"remark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.db.UpdateRepository(r.Context(), id, req.Description, req.Remark); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "repository not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "update: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "description": req.Description, "remark": req.Remark})
}

func (a *API) handleDeleteRepository(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := a.db.DeleteRepository(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "repository not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "delete: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (a *API) handleSetRepositoryTags(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req struct {
		TagIDs []int64 `json:"tag_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.db.SetRepositoryTagIDs(r.Context(), id, req.TagIDs); err != nil {
		writeError(w, http.StatusInternalServerError, "set tags: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"updated": true})
}

func (a *API) handleSyncRepository(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	repo, err := a.db.GetRepositoryByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get repository: "+err.Error())
		return
	}
	if repo == nil {
		writeError(w, http.StatusNotFound, "repository not found")
		return
	}
	if _, err := a.trk.TrackOne(r.Context(), repo.Owner, repo.Name); err != nil {
		writeError(w, http.StatusInternalServerError, "track: "+err.Error())
		return
	}
	if err := exporter.Export(r.Context(), a.db, a.exportDir); err != nil {
		writeError(w, http.StatusInternalServerError, "export: "+err.Error())
		return
	}
	updated, err := a.db.GetRepositoryByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reload: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
