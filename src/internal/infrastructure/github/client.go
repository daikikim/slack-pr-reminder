package github

import (
	"context"
	"fmt"
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
	opts := &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allPRs []model.PR

	for {
		prs, resp, err := c.client.PullRequests.List(ctx, c.owner, c.repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch PRs: %w", err)
		}

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
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

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
