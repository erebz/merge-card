package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gh "github.com/google/go-github/v86/github"

	"github.com/erebz/merge-card/internal/types"
)

// ---- NewClient ----

func TestNewClient_MissingToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")

	_, err := NewClient()
	if err != types.ErrMissingToken {
		t.Errorf("expected ErrMissingToken, got %v", err)
	}
}

func TestNewClient_MissingRepo(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "fake-token")
	t.Setenv("GITHUB_REPOSITORY", "")

	_, err := NewClient()
	if err != types.ErrMissingRepo {
		t.Errorf("expected ErrMissingRepo, got %v", err)
	}
}

func TestNewClient_MalformedRepo(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "fake-token")
	t.Setenv("GITHUB_REPOSITORY", "no-slash-here")

	_, err := NewClient()
	if err != types.ErrMissingRepo {
		t.Errorf("expected ErrMissingRepo for malformed repo, got %v", err)
	}
}

func TestNewClient_Success(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "fake-token")
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")

	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Repo() != "owner/repo" {
		t.Errorf("expected Repo() == owner/repo, got %q", c.Repo())
	}
}

// ---- buildCommentBody ----

func TestBuildCommentBody_ContainsSVG(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg"></svg>`
	body := buildCommentBody(svg)

	if !strings.Contains(body, svg) {
		t.Error("comment body does not contain the SVG")
	}
	if !strings.Contains(body, "<!-- merge-card -->") {
		t.Error("comment body missing merge-card marker comment")
	}
	if !strings.Contains(body, "<details") {
		t.Error("comment body missing <details> wrapper")
	}
}

// ---- PostComment (with a fake HTTP server) ----

// newTestClient builds a Client pointing at a fake HTTP server.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()

	httpCl := server.Client()
	ghClient := gh.NewClient(httpCl).WithAuthToken("fake-token")
	// Override the base URL to point at the test server.
	var err error
	ghClient, err = ghClient.WithEnterpriseURLs(server.URL+"/", server.URL+"/")
	if err != nil {
		t.Fatalf("set enterprise URLs: %v", err)
	}

	return &Client{gh: ghClient, owner: "owner", repo: "repo"}
}

func TestPostComment_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		comment := gh.IssueComment{ID: gh.Int64(1)}
		json.NewEncoder(w).Encode(comment) //nolint:errcheck
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.PostComment(context.Background(), 42, "<svg></svg>")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPostComment_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.PostComment(context.Background(), 42, "<svg></svg>")
	if err == nil {
		t.Error("expected error on server 500")
	}
}
