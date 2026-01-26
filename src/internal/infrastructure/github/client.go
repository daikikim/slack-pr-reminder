package github

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/dkim/slack-pr-reminder/src/internal/model"
	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// Client implements model.PRRepository.
type Client struct {
	client *github.Client
	owner  string
	repo   string
}

// NewClient creates a new GitHub client.
func NewClient(token, repository string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	parts := strings.Split(repository, "/")
	owner := parts[0]
	repo := parts[1]

	return &Client{
		client: client,
		owner:  owner,
		repo:   repo,
	}
}

// FetchOpenPRs fetches all open pull requests from the repository.
func (c *Client) FetchOpenPRs(ctx context.Context) ([]model.PR, error) {
	log.Printf("[GITHUB] Fetching open PRs from %s/%s", c.owner, c.repo)
	opts := &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allPRs []model.PR
	page := 1

	for {
		log.Printf("[GITHUB] Fetching PRs page %d", page)
		prs, resp, err := c.client.PullRequests.List(ctx, c.owner, c.repo, opts)
		if err != nil {
			log.Printf("[GITHUB] Error fetching PRs: %v", err)
			return nil, fmt.Errorf("failed to fetch PRs: %w", err)
		}

		log.Printf("[GITHUB] Received %d PRs from page %d", len(prs), page)
		for _, pr := range prs {
			var assignees []string
			for _, assignee := range pr.Assignees {
				assignees = append(assignees, assignee.GetLogin())
			}

			allPRs = append(allPRs, model.PR{
				ID:        pr.GetID(),
				Number:    pr.GetNumber(),
				Title:     pr.GetTitle(),
				URL:       pr.GetHTMLURL(),
				Author:    pr.GetUser().GetLogin(),
				Assignees: assignees,
				CreatedAt: pr.GetCreatedAt().Time,
			})
			log.Printf("[GITHUB] Added PR #%d: %s (Author: %s, Created: %s)",
				pr.GetNumber(), pr.GetTitle(), pr.GetUser().GetLogin(), pr.GetCreatedAt().Time.Format("2006-01-02 15:04:05"))
		}

		if resp.NextPage == 0 {
			log.Printf("[GITHUB] No more pages, total PRs fetched: %d", len(allPRs))
			break
		}
		opts.Page = resp.NextPage
		page++
	}

	log.Printf("[GITHUB] Successfully fetched %d open PR(s) total", len(allPRs))
	return allPRs, nil
}

// HasReviewed checks if a user has submitted a review for a PR.
func (c *Client) HasReviewed(ctx context.Context, prNumber int, username string) (bool, error) {
	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		reviews, resp, err := c.client.PullRequests.ListReviews(ctx, c.owner, c.repo, prNumber, opts)
		if err != nil {
			return false, fmt.Errorf("failed to fetch reviews: %w", err)
		}

		for _, review := range reviews {
			if review.GetUser().GetLogin() == username {
				// Check if review state is not just "PENDING"
				state := review.GetState()
				if state == "APPROVED" || state == "CHANGES_REQUESTED" || state == "COMMENTED" {
					return true, nil
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return false, nil
}

// IsMerged checks if a PR has been merged.
func (c *Client) IsMerged(ctx context.Context, prNumber int) (bool, error) {
	log.Printf("[GITHUB] Checking merge status for PR #%d in %s/%s", prNumber, c.owner, c.repo)
	isMerged, _, err := c.client.PullRequests.IsMerged(ctx, c.owner, c.repo, prNumber)
	if err != nil {
		log.Printf("[GITHUB] Error checking merge status for PR #%d: %v", prNumber, err)
		return false, fmt.Errorf("failed to check merge status for PR #%d: %w", prNumber, err)
	}
	log.Printf("[GITHUB] PR #%d merge status: %v", prNumber, isMerged)
	return isMerged, nil
}
