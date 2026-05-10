package github

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/erebz/merge-card/internal/types"
)

// testdataPath returns the absolute path to a file in the repository-level
// testdata directory, regardless of where the test binary is executed from.
func testdataPath(t *testing.T, name string) string {
	t.Helper()
	// Walk up two levels: internal/github → internal → repo root.
	abs, err := filepath.Abs(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatalf("resolve testdata path: %v", err)
	}
	return abs
}

// setEventPath sets GITHUB_EVENT_PATH and returns a cleanup function.
func setEventPath(t *testing.T, path string) {
	t.Helper()
	t.Setenv("GITHUB_EVENT_PATH", path)
}

// ---- IsMergedPR ----

func TestIsMergedPR_NilEvent(t *testing.T) {
	if IsMergedPR(nil) {
		t.Error("expected false for nil event")
	}
}

func TestIsMergedPR_MergedEvent(t *testing.T) {
	action := "closed"
	merged := true
	event := makeEvent(&action, &merged)
	if !IsMergedPR(event) {
		t.Error("expected true for merged closed event")
	}
}

func TestIsMergedPR_ClosedNotMerged(t *testing.T) {
	action := "closed"
	merged := false
	event := makeEvent(&action, &merged)
	if IsMergedPR(event) {
		t.Error("expected false for closed-but-not-merged event")
	}
}

func TestIsMergedPR_OpenedEvent(t *testing.T) {
	action := "opened"
	merged := false
	event := makeEvent(&action, &merged)
	if IsMergedPR(event) {
		t.Error("expected false for opened event")
	}
}

func TestIsMergedPR_NilAction(t *testing.T) {
	merged := true
	event := makeEvent(nil, &merged)
	if IsMergedPR(event) {
		t.Error("expected false when action is nil")
	}
}

func TestIsMergedPR_NilMergedField(t *testing.T) {
	action := "closed"
	event := makeEvent(&action, nil)
	if IsMergedPR(event) {
		t.Error("expected false when merged field is nil")
	}
}

// ---- ParseEvent ----

func TestParseEvent_MissingEnvVar(t *testing.T) {
	t.Setenv("GITHUB_EVENT_PATH", "")

	_, err := ParseEvent()
	if err == nil {
		t.Fatal("expected error when GITHUB_EVENT_PATH is empty")
	}
	if err != types.ErrMissingEventPath {
		t.Errorf("expected ErrMissingEventPath, got %v", err)
	}
}

func TestParseEvent_FileNotFound(t *testing.T) {
	t.Setenv("GITHUB_EVENT_PATH", "/nonexistent/path/event.json")

	_, err := ParseEvent()
	if err == nil {
		t.Fatal("expected error for missing event file")
	}
}

func TestParseEvent_MergedPR(t *testing.T) {
	setEventPath(t, testdataPath(t, "pull_request_merged.json"))

	pr, err := ParseEvent()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr == nil {
		t.Fatal("expected non-nil PullRequest")
	}
	if !pr.IsMerged {
		t.Error("expected IsMerged to be true")
	}
	if pr.Author != "octocat" {
		t.Errorf("expected author octocat, got %q", pr.Author)
	}
	if pr.Number != 42 {
		t.Errorf("expected PR number 42, got %d", pr.Number)
	}
	if pr.Title != "feat: add merge card generation" {
		t.Errorf("unexpected title: %q", pr.Title)
	}
}

func TestParseEvent_ClosedNotMerged(t *testing.T) {
	setEventPath(t, testdataPath(t, "pull_request_closed_not_merged.json"))

	pr, err := ParseEvent()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr.IsMerged {
		t.Error("expected IsMerged to be false for closed-not-merged PR")
	}
}

func TestParseEvent_OpenedPR(t *testing.T) {
	setEventPath(t, testdataPath(t, "pull_request_opened.json"))

	pr, err := ParseEvent()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr.IsMerged {
		t.Error("expected IsMerged to be false for opened PR")
	}
}

func TestParseEvent_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "bad.json")
	if err := os.WriteFile(f, []byte("{not valid json}"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_EVENT_PATH", f)

	_, err := ParseEvent()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseEvent_NoPullRequest(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "no_pr.json")
	// Valid JSON but no pull_request key.
	if err := os.WriteFile(f, []byte(`{"action":"closed"}`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_EVENT_PATH", f)

	_, err := ParseEvent()
	if err != types.ErrNilPullRequest {
		t.Errorf("expected ErrNilPullRequest, got %v", err)
	}
}
