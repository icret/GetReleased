package api

import (
	"net/http"

	"getreleased/internal/database"
)

type statsResponse struct {
	Overview        database.StatsOverview       `json:"overview"`
	Languages       []database.LanguageCount     `json:"languages"`
	TopRepositories []database.TopRepository     `json:"top_repositories"`
	RecentReleases  []database.RecentRelease     `json:"recent_releases"`
	ReleaseTrend    []database.ReleaseTrendPoint `json:"release_trend"`
	TagTypes        []database.TagTypeCount      `json:"tag_types"`
}

func (a *API) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := statsResponse{}

	overview, err := a.db.StatsOverview(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "stats overview: "+err.Error())
		return
	}
	resp.Overview = overview

	if resp.Languages, err = a.db.StatsLanguages(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, "stats languages: "+err.Error())
		return
	}
	if resp.TopRepositories, err = a.db.StatsTopRepositories(ctx, 10); err != nil {
		writeError(w, http.StatusInternalServerError, "stats top repos: "+err.Error())
		return
	}
	if resp.RecentReleases, err = a.db.StatsRecentReleases(ctx, 10); err != nil {
		writeError(w, http.StatusInternalServerError, "stats recent releases: "+err.Error())
		return
	}
	if resp.ReleaseTrend, err = a.db.StatsReleaseTrend(ctx, 12); err != nil {
		writeError(w, http.StatusInternalServerError, "stats trend: "+err.Error())
		return
	}
	if resp.TagTypes, err = a.db.StatsTagTypes(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, "stats tag types: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
