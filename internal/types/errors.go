package types

import "errors"

// Sentinel errors returned by merge-card packages.
var (
	// ErrNotMerged is returned when the pull request event is not a merge.
	ErrNotMerged = errors.New("pull request was not merged")

	// ErrMissingToken is returned when GITHUB_TOKEN is not set in the environment.
	ErrMissingToken = errors.New("GITHUB_TOKEN environment variable is not set")

	// ErrMissingRepo is returned when GITHUB_REPOSITORY is not set or malformed.
	ErrMissingRepo = errors.New("GITHUB_REPOSITORY environment variable is missing or invalid")

	// ErrMissingEventPath is returned when GITHUB_EVENT_PATH is not set.
	ErrMissingEventPath = errors.New("GITHUB_EVENT_PATH environment variable is not set")

	// ErrNilPullRequest is returned when the event payload contains no pull request.
	ErrNilPullRequest = errors.New("pull request is nil in event payload")
)
