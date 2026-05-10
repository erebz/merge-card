// merge-card is the entry point for the GitHub Action.
//
// Execution flow:
//  1. Parse the GitHub Actions pull_request event from GITHUB_EVENT_PATH.
//  2. If the PR is not merged, exit cleanly without posting a comment.
//  3. Build a GitHub API client authenticated via GITHUB_TOKEN.
//  4. Generate the SVG reward card for the pull request author.
//  5. Post the card as a comment on the pull request.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/erebz/merge-card/internal/cards"
	"github.com/erebz/merge-card/internal/github"
	"github.com/erebz/merge-card/internal/types"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("merge-card: %v", err)
	}
}

// run contains the full action logic, separated from main() to allow clean
// error propagation and easier unit testing in the future.
func run() error {
	log.Println("merge-card: starting")

	// Step 1 – parse the pull request event.
	pr, err := github.ParseEvent()
	if err != nil {
		// A missing event path means we're not running inside GitHub Actions;
		// treat it as a hard failure.
		if errors.Is(err, types.ErrMissingEventPath) {
			return fmt.Errorf("not running inside a GitHub Actions environment: %w", err)
		}
		return fmt.Errorf("parse event: %w", err)
	}

	// Step 2 – skip non-merge events silently.
	if !pr.IsMerged {
		log.Println("merge-card: pull request was not merged — skipping")
		os.Exit(0)
	}

	log.Printf("merge-card: merged PR #%d by %s — generating card", pr.Number, pr.Author)

	// Step 3 – build the GitHub API client.
	client, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("create GitHub client: %w", err)
	}

	// Step 4 – generate the SVG card.
	svg, err := cards.Generate(cards.Data{
		Author: pr.Author,
		Title:  pr.Title,
		Number: pr.Number,
		Repo:   client.Repo(),
	})
	if err != nil {
		return fmt.Errorf("generate card: %w", err)
	}

	// Step 5 – post the card as a PR comment.
	if err := client.PostComment(context.Background(), pr.Number, svg); err != nil {
		return fmt.Errorf("post comment: %w", err)
	}

	log.Printf("merge-card: card posted on PR #%d", pr.Number)
	return nil
}
