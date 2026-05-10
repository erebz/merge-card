package github

import gh "github.com/google/go-github/v86/github"

// makeEvent builds a minimal PullRequestEvent for use in table-driven tests.
// Pass nil for action or merged to leave those fields unset.
func makeEvent(action *string, merged *bool) *gh.PullRequestEvent {
	pr := &gh.PullRequest{}
	if merged != nil {
		pr.Merged = merged
	}
	e := &gh.PullRequestEvent{
		Action:      action,
		PullRequest: pr,
	}
	return e
}
