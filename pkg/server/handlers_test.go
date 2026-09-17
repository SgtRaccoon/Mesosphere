package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mesosphere/mesosphere/pkg/core/config"
	"github.com/mesosphere/mesosphere/pkg/core/docs"
	"github.com/mesosphere/mesosphere/pkg/core/repo"
	"github.com/mesosphere/mesosphere/pkg/core/tasks"
	"github.com/mesosphere/mesosphere/pkg/git"
)

func withAPI(t *testing.T, a *API) http.Handler {
	t.Helper()
	prev := activeAPI
	activeAPI = a
	t.Cleanup(func() { activeAPI = prev })
	return NewRouter()
}

func TestConfigAndReposEndpoints(t *testing.T) {
	cfg := &config.Config{
		Version: "1.0",
		Repositories: []config.RepositoryConfig{
			{ID: "main", Name: "Main", Path: "/tmp/main"},
		},
	}
	var saved *config.Config
	h := withAPI(t, &API{
		LoadConfig: func() (*config.Config, error) { return cfg, nil },
		SaveConfig: func(c *config.Config) error { saved = c; cfg = c; return nil },
		RepoStatus: func(repoPath string) (*repo.RepoStatus, error) {
			return &repo.RepoStatus{HasUncommitted: true, Behind: 2, Ahead: 0, Branch: "main"}, nil
		},
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/config", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET config %d %s", rec.Code, rec.Body.String())
	}
	var got config.Config
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Repositories) != 1 {
		t.Fatalf("repos in config: %+v", got)
	}

	body, _ := json.Marshal(config.Config{Version: "1.0", Repositories: []config.RepositoryConfig{
		{ID: "main", Name: "Main", Path: "/tmp/main"},
		{ID: "other", Name: "Other", Path: "/tmp/other"},
	}})
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/config", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST config %d %s", rec.Code, rec.Body.String())
	}
	if saved == nil || len(saved.Repositories) != 2 {
		t.Fatalf("saved = %+v", saved)
	}

	rec = httptest.NewRecorder()
	addBody, _ := json.Marshal(config.RepositoryConfig{Name: "Third", Path: "/tmp/third"})
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/repos", bytes.NewReader(addBody)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST repos %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/repos", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET repos %d", rec.Code)
	}
	var list []config.RepositoryConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("list = %+v", list)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/repos/main/status", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var st repo.RepoStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
		t.Fatal(err)
	}
	if !st.HasUncommitted || st.Behind != 2 {
		t.Fatalf("status = %+v", st)
	}
}

func TestRepoStatusNotFound(t *testing.T) {
	h := withAPI(t, &API{
		LoadConfig: func() (*config.Config, error) {
			return &config.Config{}, nil
		},
		RepoStatus: func(string) (*repo.RepoStatus, error) {
			return nil, errors.New("unused")
		},
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/repos/missing/status", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d", rec.Code)
	}
}

func TestSyncAndPublish(t *testing.T) {
	cfg := &config.Config{Repositories: []config.RepositoryConfig{
		{ID: "main", Path: "/tmp/main"},
	}}
	synced, published := false, false
	h := withAPI(t, &API{
		LoadConfig: func() (*config.Config, error) { return cfg, nil },
		Sync: func(path string) error {
			synced = true
			return nil
		},
		Publish: func(path string) error {
			published = true
			return nil
		},
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/repos/main/sync", nil))
	if rec.Code != http.StatusOK || !synced {
		t.Fatalf("sync %d synced=%v %s", rec.Code, synced, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/repos/main/publish", nil))
	if rec.Code != http.StatusOK || !published {
		t.Fatalf("publish %d", rec.Code)
	}
}

func TestPublishConflict(t *testing.T) {
	cfg := &config.Config{Repositories: []config.RepositoryConfig{
		{ID: "main", Path: "/tmp/main"},
	}}
	h := withAPI(t, &API{
		LoadConfig: func() (*config.Config, error) { return cfg, nil },
		Publish: func(string) error {
			return &git.ConflictError{Op: "pull", Output: "CONFLICT"}
		},
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/repos/main/publish", nil))
	if rec.Code != http.StatusConflict {
		t.Fatalf("code = %d %s", rec.Code, rec.Body.String())
	}
}

func TestDocumentEndpoints(t *testing.T) {
	cfg := &config.Config{Repositories: []config.RepositoryConfig{
		{ID: "main", Path: "/tmp/main"},
	}}
	saved := false
	h := withAPI(t, &API{
		LoadConfig: func() (*config.Config, error) { return cfg, nil },
		ListDocs: func(string) ([]docs.DocumentSummary, error) {
			return []docs.DocumentSummary{{Path: "readme.md", Title: "Hello"}}, nil
		},
		GetDoc: func(repoPath, docPath, version string) (*docs.DocumentDetail, error) {
			return &docs.DocumentDetail{Path: docPath, Version: version, Content: "body"}, nil
		},
		ListDocVer: func(string, string) ([]git.CommitInfo, error) {
			return []git.CommitInfo{{Hash: "abc", Message: "first"}}, nil
		},
		SaveDoc: func(repoPath, docPath, content, commitMsg string) error {
			saved = true
			return nil
		},
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/repos/main/docs", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/repos/main/docs/detail?path=readme.md&version=abc", nil))
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte("body")) {
		t.Fatalf("detail %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/repos/main/docs/versions?path=readme.md", nil))
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte("abc")) {
		t.Fatalf("versions %d %s", rec.Code, rec.Body.String())
	}

	body, _ := json.Marshal(saveDocRequest{Path: "readme.md", Content: "x", CommitMsg: "m"})
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/repos/main/docs/save", bytes.NewReader(body)))
	if rec.Code != http.StatusOK || !saved {
		t.Fatalf("save %d saved=%v %s", rec.Code, saved, rec.Body.String())
	}
}

func TestTaskEndpoints(t *testing.T) {
	cfg := &config.Config{Repositories: []config.RepositoryConfig{
		{ID: "main", Path: "/tmp/main"},
	}}
	saved := false
	sample := tasks.CommonTask{CardID: "TASK-101", Title: "Ship", RawContent: `{"id":"TASK-101"}`}
	h := withAPI(t, &API{
		LoadConfig: func() (*config.Config, error) { return cfg, nil },
		ListTasks: func(string, string) ([]tasks.CommonTask, error) {
			return []tasks.CommonTask{sample}, nil
		},
		GetTask: func(_, _, id string) (*tasks.CommonTask, error) {
			if id != "TASK-101" {
				return nil, errors.New("task not found: " + id)
			}
			return &sample, nil
		},
		SaveTask: func(repoPath, taskID, rawJSON, commitMsg string) error {
			saved = true
			return nil
		},
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/repos/main/tasks", nil))
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte("TASK-101")) {
		t.Fatalf("list %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/repos/main/tasks/TASK-101", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("get %d %s", rec.Code, rec.Body.String())
	}

	body, _ := json.Marshal(saveTaskRequest{RawJSON: `{"id":"TASK-101"}`, CommitMsg: "m"})
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/repos/main/tasks/TASK-101/save", bytes.NewReader(body)))
	if rec.Code != http.StatusOK || !saved {
		t.Fatalf("save %d saved=%v %s", rec.Code, saved, rec.Body.String())
	}
}
