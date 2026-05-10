package github

import (
	"encoding/json"
	"os"

	gh "github.com/google/go-github/v86/github"

	"github.com/erebz/merge-card/internal/types"
)

// ParseEvent reads the GitHub Actions event file pointed to by
// GITHUB_EVENT_PATH, unmarshals it as a PullRequestEvent, and returns a
// *types.PullRequest populated from the event payload.
//
// It returns:
//   - types.ErrMissingEventPath when GITHUB_EVENT_PATH is not set.
//   - types.ErrNotMerged when the event is a pull request that was not merged.
//   - types.ErrNilPullRequest when the event payload contains no pull request.
//   - A JSON/IO error when the event file cannot be read or decoded.
func ParseEvent() (*types.PullRequest, error) {
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return nil, types.ErrMissingEventPath
	}

	data, err := os.ReadFile(eventPath)
	if err != nil {
		return nil, err
	}

	var prEvent gh.PullRequestEvent
	if err = json.Unmarshal(data, &prEvent); err != nil {
		return nil, err
	}

	if prEvent.PullRequest == nil {
		return nil, types.ErrNilPullRequest
	}

	merged := IsMergedPR(&prEvent)

	return &types.PullRequest{
		Author:   prEvent.PullRequest.User.GetLogin(),
		Title:    prEvent.PullRequest.GetTitle(),
		Number:   prEvent.PullRequest.GetNumber(),
		IsMerged: merged,
	}, nil
}

// IsMergedPR reports whether the given PullRequestEvent represents a
// successfully merged pull request (action == "closed" and merged == true).
func IsMergedPR(e *gh.PullRequestEvent) bool {
	if e == nil || e.Action == nil || e.PullRequest == nil {
		return false
	}
	return *e.Action == "closed" &&
		e.PullRequest.Merged != nil &&
		*e.PullRequest.Merged
}
