package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mesosphere/mesosphere/pkg/core/repo"
	"github.com/mesosphere/mesosphere/pkg/git"
)

func (a *API) lookupRepoPath(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := chi.URLParam(r, "repoID")
	cfg, err := a.LoadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return "", false
	}
	found, err := repo.GetRepositoryContext(cfg, "", id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return "", false
	}
	return found.Path, true
}

func gitStatus(err error) int {
	var ce *git.ConflictError
	if errors.As(err, &ce) {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

func (a *API) handleSync(w http.ResponseWriter, r *http.Request) {
	path, ok := a.lookupRepoPath(w, r)
	if !ok {
		return
	}
	if err := a.Sync(path); err != nil {
		writeError(w, gitStatus(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "synced"})
}

func (a *API) handlePublish(w http.ResponseWriter, r *http.Request) {
	path, ok := a.lookupRepoPath(w, r)
	if !ok {
		return
	}
	if err := a.Publish(path); err != nil {
		writeError(w, gitStatus(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
}

func (a *API) mountSync(r chi.Router) {
	r.Post("/repos/{repoID}/sync", a.handleSync)
	r.Post("/repos/{repoID}/publish", a.handlePublish)
}
