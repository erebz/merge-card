// Package github provides helpers for interacting with the GitHub API inside
// a GitHub Actions environment.
//
// Authentication uses the GITHUB_TOKEN environment variable, which is
// automatically provided by the Actions runner when the workflow maps it.
package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	gh "github.com/google/go-github/v86/github"
	"golang.org/x/oauth2"

	"github.com/erebz/merge-card/internal/types"
)

// Client wraps the go-github client and exposes high-level operations used by
// the action.
type Client struct {
	gh *gh.Client
	// owner and repo are derived from GITHUB_REPOSITORY ("owner/repo").
	owner string
	repo  string
}

// NewClient constructs a Client authenticated with the token stored in the
// GITHUB_TOKEN environment variable. GITHUB_REPOSITORY must also be set and
// must be in "owner/repo" format.
//
// It returns types.ErrMissingToken when GITHUB_TOKEN is absent and
// types.ErrMissingRepo when GITHUB_REPOSITORY is absent or malformed.
func NewClient() (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, types.ErrMissingToken
	}

	repoEnv := os.Getenv("GITHUB_REPOSITORY")
	if repoEnv == "" {
		return nil, types.ErrMissingRepo
	}

	parts := strings.SplitN(repoEnv, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, types.ErrMissingRepo
	}

	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	return &Client{
		gh:    gh.NewClient(httpClient),
		owner: parts[0],
		repo:  parts[1],
	}, nil
}

// Repo returns the "owner/repo" string, e.g. "octocat/hello-world".
func (c *Client) Repo() string {
	return c.owner + "/" + c.repo
}

// PostComment creates an issue comment on the pull request identified by
// prNumber. In GitHub's API, PR comments are created via the Issues endpoint.
//
// The svgCard string should be a fully-rendered SVG. It is wrapped in an HTML
// <details> block so the card is collapsible and the raw SVG does not clutter
// the PR timeline.
func (c *Client) PostComment(ctx context.Context, prNumber int, svgCard string) error {
	body := buildCommentBody(svgCard)

	_, resp, err := c.gh.Issues.CreateComment(ctx, c.owner, c.repo, prNumber, &gh.IssueComment{
		Body: gh.String(body),
	})
	if err != nil {
		return fmt.Errorf("create comment on PR #%d: %w", prNumber, err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status %d when commenting on PR #%d", resp.StatusCode, prNumber)
	}
	return nil
}

// buildCommentBody wraps the SVG in a markdown-compatible HTML block that
// renders inline on GitHub.
func buildCommentBody(svgCard string) string {
	return fmt.Sprintf(
		"<!-- merge-card -->\n"+
			"<details open>\n"+
			"<summary>🎉 Merge Card</summary>\n\n"+
			"%s\n\n"+
			"</details>",
		svgCard,
	)
}
