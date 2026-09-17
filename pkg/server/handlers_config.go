package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mesosphere/mesosphere/pkg/core/config"
	"github.com/mesosphere/mesosphere/pkg/core/docs"
	"github.com/mesosphere/mesosphere/pkg/core/repo"
	"github.com/mesosphere/mesosphere/pkg/core/tasks"
	"github.com/mesosphere/mesosphere/pkg/git"
)

// API hosts injectable core operations for HTTP handlers.
type API struct {
	LoadConfig func() (*config.Config, error)
	SaveConfig func(cfg *config.Config) error
	RepoStatus func(repoPath string) (*repo.RepoStatus, error)
	Sync       func(repoPath string) error
	Publish    func(repoPath string) error
	ListDocs   func(repoPath string) ([]docs.DocumentSummary, error)
	GetDoc     func(repoPath, docPath, version string) (*docs.DocumentDetail, error)
	ListDocVer func(repoPath, docPath string) ([]git.CommitInfo, error)
	SaveDoc    func(repoPath, docPath, content, commitMsg string) error
	ListTasks  func(repoPath, repoID string) ([]tasks.CommonTask, error)
	GetTask    func(repoPath, repoID, taskID string) (*tasks.CommonTask, error)
	SaveTask   func(repoPath, taskID, rawJSON, commitMsg string) error
}

func defaultAPI() *API {
	return &API{
		LoadConfig: func() (*config.Config, error) {
			return config.LoadConfig("")
		},
		SaveConfig: func(cfg *config.Config) error {
			return config.SaveConfig("", cfg)
		},
		RepoStatus: repo.GetRepoStatus,
		Sync:       repo.SyncRepository,
		Publish:    repo.PublishRepository,
		ListDocs:   defaultDocs.ListDocuments,
		GetDoc:     defaultDocs.GetDocument,
		ListDocVer: defaultDocs.ListDocumentVersions,
		SaveDoc:    defaultDocs.SaveDocument,
		ListTasks:  tasks.ListTasks,
		GetTask:    tasks.GetTask,
		SaveTask:   tasks.SaveTask,
	}
}

var defaultDocs = &docs.Engine{}

var activeAPI = defaultAPI()

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func (a *API) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := a.LoadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (a *API) handlePostConfig(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := a.SaveConfig(&cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, &cfg)
}

func (a *API) handleListRepos(w http.ResponseWriter, r *http.Request) {
	cfg, err := a.LoadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, repo.ListRepositories(cfg))
}

func (a *API) handleRepoStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "repoID")
	cfg, err := a.LoadConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	found, err := repo.GetRepositoryContext(cfg, "", id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	st, err := a.RepoStatus(found.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (a *API) mountConfig(r chi.Router) {
	r.Get("/config", a.handleGetConfig)
	r.Post("/config", a.handlePostConfig)
	r.Get("/repos", a.handleListRepos)
	r.Get("/repos/{repoID}/status", a.handleRepoStatus)
}
