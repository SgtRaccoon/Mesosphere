package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *API) handleListDocs(w http.ResponseWriter, r *http.Request) {
	path, ok := a.lookupRepoPath(w, r)
	if !ok {
		return
	}
	list, err := a.ListDocs(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *API) handleDocDetail(w http.ResponseWriter, r *http.Request) {
	path, ok := a.lookupRepoPath(w, r)
	if !ok {
		return
	}
	docPath := r.URL.Query().Get("path")
	if docPath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "path is required"})
		return
	}
	doc, err := a.GetDoc(path, docPath, r.URL.Query().Get("version"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (a *API) handleDocVersions(w http.ResponseWriter, r *http.Request) {
	path, ok := a.lookupRepoPath(w, r)
	if !ok {
		return
	}
	docPath := r.URL.Query().Get("path")
	if docPath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "path is required"})
		return
	}
	vers, err := a.ListDocVer(path, docPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, vers)
}

type saveDocRequest struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	CommitMsg string `json:"commit_msg"`
}

func (a *API) handleSaveDoc(w http.ResponseWriter, r *http.Request) {
	path, ok := a.lookupRepoPath(w, r)
	if !ok {
		return
	}
	var req saveDocRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Path == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "path is required"})
		return
	}
	if err := a.SaveDoc(path, req.Path, req.Content, req.CommitMsg); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *API) mountDocs(r chi.Router) {
	r.Get("/repos/{repoID}/docs", a.handleListDocs)
	r.Get("/repos/{repoID}/docs/detail", a.handleDocDetail)
	r.Get("/repos/{repoID}/docs/versions", a.handleDocVersions)
	r.Post("/repos/{repoID}/docs/save", a.handleSaveDoc)
}
