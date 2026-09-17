package tasks

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// FrameworkTranslator converts a native task file into CommonTask values.
type FrameworkTranslator interface {
	Name() string
	Detect(path, content string) bool
	Translate(path, content, repoID string) ([]CommonTask, error)
}

var defaultTranslators = []FrameworkTranslator{
	githubSpecKitTranslator{},
	openSpecTranslator{},
	bmadTranslator{},
	kiroTranslator{},
	cosmosTranslator{},
	jsonTaskTranslator{},
	markdownCheckboxTranslator{},
}

var checkboxRE = regexp.MustCompile(`(?m)^\s*[-*]\s*\[([ xX])\]\s+(.*)$`)

// ListTasks scans repoPath for supported task files and translates them.
func ListTasks(repoPath, repoID string) ([]CommonTask, error) {
	return ListTasksWith(repoPath, repoID, defaultTranslators)
}

// ListTasksWith is ListTasks with an explicit translator set.
func ListTasksWith(repoPath, repoID string, translators []FrameworkTranslator) ([]CommonTask, error) {
	var out []CommonTask
	err := filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext != ".md" && ext != ".json" && ext != ".yaml" && ext != ".yml" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)
		rel, err := filepath.Rel(repoPath, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		tr := detect(translators, rel, content)
		if tr == nil {
			return nil
		}
		tasks, err := tr.Translate(rel, content, repoID)
		if err != nil {
			return nil
		}
		out = append(out, tasks...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []CommonTask{}
	}
	return out, nil
}

// GetTask returns the first task whose CardID matches.
func GetTask(repoPath, repoID, taskID string) (*CommonTask, error) {
	all, err := ListTasks(repoPath, repoID)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].CardID == taskID {
			t := all[i]
			return &t, nil
		}
	}
	return nil, fmt.Errorf("task not found: %s", taskID)
}

func skipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", "dist":
		return true
	default:
		return false
	}
}

func detect(translators []FrameworkTranslator, path, content string) FrameworkTranslator {
	for _, t := range translators {
		if t.Detect(path, content) {
			return t
		}
	}
	return nil
}

func baseTask(path, repoID, raw, framework string) CommonTask {
	return CommonTask{
		FilePath:    path,
		RepoID:      repoID,
		RawContent:  raw,
		Framework:   framework,
		BlockerID:   []string{},
		ExtraFields: map[string]interface{}{},
		Status:      StatusTodo,
	}
}

func strField(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if t != "" {
					return t
				}
			case fmt.Stringer:
				s := t.String()
				if s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func boolField(m map[string]interface{}, keys ...string) bool {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
	}
	return false
}

func stringSlice(v interface{}) []string {
	switch t := v.(type) {
	case nil:
		return []string{}
	case []string:
		return t
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, x := range t {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		if t == "" {
			return []string{}
		}
		return []string{t}
	default:
		return []string{}
	}
}

func asObject(v interface{}) (map[string]interface{}, bool) {
	m, ok := v.(map[string]interface{})
	return m, ok
}

func parseLoose(content string) (interface{}, error) {
	trim := strings.TrimSpace(content)
	if trim == "" {
		return nil, fmt.Errorf("empty")
	}
	if trim[0] == '{' || trim[0] == '[' {
		var v interface{}
		if err := json.Unmarshal([]byte(trim), &v); err == nil {
			return v, nil
		}
	}
	var v interface{}
	if err := yaml.Unmarshal([]byte(content), &v); err != nil {
		return nil, err
	}
	return v, nil
}

func mapFromUnknown(v interface{}) map[string]interface{} {
	if m, ok := asObject(v); ok {
		return m
	}
	return nil
}

func applyCommonMap(t *CommonTask, m map[string]interface{}) {
	t.CardID = firstNonEmpty(t.CardID, strField(m, "card_id", "id", "task_id", "key"))
	t.Title = firstNonEmpty(t.Title, strField(m, "title", "name", "summary"))
	t.Scope = firstNonEmpty(t.Scope, strField(m, "scope", "description", "body"))
	t.ValidationGate = firstNonEmpty(t.ValidationGate, strField(m, "validation_gate", "acceptance", "acceptanceCriteria", "validation"))
	if s := strField(m, "status", "state", "column"); s != "" {
		t.Status = normalizeStatus(s)
	} else {
		t.Status = normalizeStatus(t.Status)
	}
	if boolField(m, "is_blocked", "blocked") {
		t.IsBlocked = true
	}
	if ids := stringSlice(m["blocker_ids"]); len(ids) > 0 {
		t.BlockerID = ids
	} else if ids := stringSlice(m["blockers"]); len(ids) > 0 {
		t.BlockerID = ids
	} else if s := strField(m, "blocker_id", "blocked_by"); s != "" {
		t.BlockerID = []string{s}
	}
	if t.IsBlocked && t.Status == StatusDone {
		t.Status = StatusTodo
	}
	t.ExtraFields = m
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func normalizeStatus(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "todo", "to do", "to-do", "open", "backlog", "pending":
		return StatusTodo
	case "in progress", "in-progress", "in_progress", "doing", "wip", "started":
		return StatusInProgress
	case "done", "complete", "completed", "closed":
		return StatusDone
	default:
		if s == "" {
			return StatusTodo
		}
		return s
	}
}

func objectsFromRoot(v interface{}, keys ...string) []map[string]interface{} {
	if m, ok := asObject(v); ok {
		for _, k := range keys {
			if inner, ok := m[k]; ok {
				return objectsFromRoot(inner)
			}
		}
		return []map[string]interface{}{m}
	}
	if arr, ok := v.([]interface{}); ok {
		var out []map[string]interface{}
		for _, item := range arr {
			if m, ok := asObject(item); ok {
				out = append(out, m)
			}
		}
		return out
	}
	return nil
}

type githubSpecKitTranslator struct{}

func (githubSpecKitTranslator) Name() string { return "github-spec-kit" }
func (githubSpecKitTranslator) Detect(path, content string) bool {
	p := strings.ToLower(path)
	if strings.Contains(p, ".specify") || strings.Contains(p, "spec-kit") {
		return true
	}
	return strings.Contains(content, `"speckit"`) || strings.Contains(content, "spec-kit")
}
func (t githubSpecKitTranslator) Translate(path, content, repoID string) ([]CommonTask, error) {
	return translateMapped(t.Name(), path, content, repoID, "tasks", "items")
}

type openSpecTranslator struct{}

func (openSpecTranslator) Name() string { return "openspec" }
func (openSpecTranslator) Detect(path, content string) bool {
	p := strings.ToLower(path)
	return strings.Contains(p, "openspec") || strings.Contains(content, "openspec")
}
func (t openSpecTranslator) Translate(path, content, repoID string) ([]CommonTask, error) {
	return translateMapped(t.Name(), path, content, repoID, "tasks", "changes")
}

type bmadTranslator struct{}

func (bmadTranslator) Name() string { return "bmad-method" }
func (bmadTranslator) Detect(path, content string) bool {
	p := strings.ToLower(path)
	return strings.Contains(p, "bmad") || strings.Contains(strings.ToLower(content), "bmad")
}
func (t bmadTranslator) Translate(path, content, repoID string) ([]CommonTask, error) {
	return translateMapped(t.Name(), path, content, repoID, "stories", "tasks")
}

type kiroTranslator struct{}

func (kiroTranslator) Name() string { return "amazon-kiro" }
func (kiroTranslator) Detect(path, content string) bool {
	p := strings.ToLower(path)
	return strings.Contains(p, ".kiro") || strings.Contains(p, "kiro") || strings.Contains(strings.ToLower(content), `"kiro"`)
}
func (t kiroTranslator) Translate(path, content, repoID string) ([]CommonTask, error) {
	return translateMapped(t.Name(), path, content, repoID, "tasks")
}

type cosmosTranslator struct{}

func (cosmosTranslator) Name() string { return "augment-cosmos" }
func (cosmosTranslator) Detect(path, content string) bool {
	p := strings.ToLower(path)
	return strings.Contains(p, "cosmos") || strings.Contains(strings.ToLower(content), "augment-cosmos")
}
func (t cosmosTranslator) Translate(path, content, repoID string) ([]CommonTask, error) {
	return translateMapped(t.Name(), path, content, repoID, "tasks", "cards")
}

type jsonTaskTranslator struct{}

func (jsonTaskTranslator) Name() string { return "json" }
func (jsonTaskTranslator) Detect(path, content string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".json" && ext != ".yaml" && ext != ".yml" {
		return false
	}
	v, err := parseLoose(content)
	if err != nil {
		return false
	}
	objs := objectsFromRoot(v, "tasks", "items")
	if len(objs) == 0 {
		return false
	}
	m := objs[0]
	return strField(m, "id", "card_id", "title") != ""
}
func (t jsonTaskTranslator) Translate(path, content, repoID string) ([]CommonTask, error) {
	return translateMapped(t.Name(), path, content, repoID, "tasks", "items")
}

type markdownCheckboxTranslator struct{}

func (markdownCheckboxTranslator) Name() string { return "markdown-checkbox" }
func (markdownCheckboxTranslator) Detect(path, content string) bool {
	if strings.ToLower(filepath.Ext(path)) != ".md" {
		return false
	}
	return checkboxRE.MatchString(content)
}
func (t markdownCheckboxTranslator) Translate(path, content, repoID string) ([]CommonTask, error) {
	matches := checkboxRE.FindAllStringSubmatch(content, -1)
	var out []CommonTask
	for i, m := range matches {
		ct := baseTask(path, repoID, m[0], t.Name())
		checked := strings.EqualFold(m[1], "x")
		if checked {
			ct.Status = StatusDone
		} else {
			ct.Status = StatusTodo
		}
		title := strings.TrimSpace(m[2])
		ct.Title = title
		ct.CardID = fmt.Sprintf("%s#%d", path, i+1)
		out = append(out, ct)
	}
	return out, nil
}

func translateMapped(framework, path, content, repoID string, nestKeys ...string) ([]CommonTask, error) {
	v, err := parseLoose(content)
	if err != nil {
		return nil, err
	}
	objs := objectsFromRoot(v, nestKeys...)
	if len(objs) == 0 {
		return []CommonTask{}, nil
	}
	raw := content
	var out []CommonTask
	for i, m := range objs {
		ct := baseTask(path, repoID, raw, framework)
		applyCommonMap(&ct, m)
		if ct.CardID == "" {
			ct.CardID = fmt.Sprintf("%s#%d", path, i+1)
		}
		if ct.Title == "" {
			ct.Title = ct.CardID
		}
		out = append(out, ct)
	}
	return out, nil
}
