package types

type PullRequest struct {
	Author   string
	Title    string
	Number   int
	IsMerged bool
}
