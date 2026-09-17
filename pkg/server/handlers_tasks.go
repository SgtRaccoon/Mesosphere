package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (a *API) handleListTasks(w http.ResponseWriter, r *http.Request) {
	path, ok := a.lookupRepoPath(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "repoID")
	list, err := a.ListTasks(path, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *API) handleGetTask(w http.ResponseWriter, r *http.Request) {
	path, ok := a.lookupRepoPath(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "repoID")
	taskID := chi.URLParam(r, "taskID")
	task, err := a.GetTask(path, id, taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

type saveTaskRequest struct {
	RawJSON   string `json:"raw_json"`
	CommitMsg string `json:"commit_msg"`
}

func (a *API) handleSaveTask(w http.ResponseWriter, r *http.Request) {
	path, ok := a.lookupRepoPath(w, r)
	if !ok {
		return
	}
	taskID := chi.URLParam(r, "taskID")
	var req saveTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := a.SaveTask(path, taskID, req.RawJSON, req.CommitMsg); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) mountTasks(r chi.Router) {
	r.Get("/repos/{repoID}/tasks", a.handleListTasks)
	r.Get("/repos/{repoID}/tasks/{taskID}", a.handleGetTask)
	r.Post("/repos/{repoID}/tasks/{taskID}/save", a.handleSaveTask)
}
