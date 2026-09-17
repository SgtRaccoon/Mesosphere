// Package tasks translates framework-specific task files into a common representation.
package tasks

// CommonTask is the canonical internal task schema (DESIGN §5.2).
type CommonTask struct {
	CardID         string                 `json:"card_id"`
	Title          string                 `json:"title"`
	Scope          string                 `json:"scope"`
	ValidationGate string                 `json:"validation_gate"`
	IsBlocked      bool                   `json:"is_blocked"`
	BlockerID      []string               `json:"blocker_ids"`
	Status         string                 `json:"status"`
	RawContent     string                 `json:"raw_content"`
	FilePath       string                 `json:"file_path"`
	RepoID         string                 `json:"repo_id"`
	ExtraFields    map[string]interface{} `json:"extra_fields"`
	Framework      string                 `json:"framework,omitempty"`
}

const (
	StatusTodo       = "TO DO"
	StatusInProgress = "IN PROGRESS"
	StatusDone       = "DONE"
)

// DeriveColumns implements DESIGN §5.3.
func DeriveColumns(tasks []CommonTask, binarySchema bool) []string {
	if binarySchema {
		return []string{StatusTodo, StatusDone}
	}
	seen := map[string]struct{}{}
	var cols []string
	for _, t := range tasks {
		st := t.Status
		if st == "" {
			st = StatusTodo
		}
		if _, ok := seen[st]; ok {
			continue
		}
		seen[st] = struct{}{}
		cols = append(cols, st)
	}
	if len(cols) == 0 {
		return []string{StatusTodo}
	}
	return cols
}
