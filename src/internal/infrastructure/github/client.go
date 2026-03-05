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

			var requestedReviewers []string
			for _, reviewer := range pr.RequestedReviewers {
				requestedReviewers = append(requestedReviewers, reviewer.GetLogin())
			}

			allPRs = append(allPRs, model.PR{
				ID:                 pr.GetID(),
				Number:             pr.GetNumber(),
				Title:              pr.GetTitle(),
				URL:                pr.GetHTMLURL(),
				Author:             pr.GetUser().GetLogin(),
				Assignees:          assignees,
				RequestedReviewers: requestedReviewers,
				CreatedAt:          pr.GetCreatedAt().Time,
			})
			log.Printf("[GITHUB] Added PR #%d: %s (Author: %s, Assignees: %d, RequestedReviewers: %d, Created: %s)",
				pr.GetNumber(),
				pr.GetTitle(),
				pr.GetUser().GetLogin(),
				len(assignees),
				len(requestedReviewers),
				pr.GetCreatedAt().Time.Format("2006-01-02 15:04:05"),
			)
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

// GetReviewStatuses returns a map of username to their latest review state.
func (c *Client) GetReviewStatuses(ctx context.Context, prNumber int) (map[string]string, error) {
	opts := &github.ListOptions{
		PerPage: 100,
	}

	reviewsByAuthor := make(map[string]string)

	for {
		reviews, resp, err := c.client.PullRequests.ListReviews(ctx, c.owner, c.repo, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch reviews: %w", err)
		}

		// Process reviews in order (ListReviews returns them in chronological order by default)
		for _, review := range reviews {
			user := review.GetUser().GetLogin()
			state := review.GetState()

			// "DISMISSED" reviews should probably be ignored or handled, but for now we just track the latest state.
			// However, usually we care about APPROVED, CHANGES_REQUESTED, COMMENTED.
			// If a user approves, then comments, they are still approved usually, BUT GitHub API returns distinct reviews.
			// Actually, if you request changes, then approve, the latest one stands.
			// If you approve, then comment, the approval stands? GitHub API is a bit complex.
			// But usually relying on the latest Review state is a good approximation.
			// Note: "COMMENTED" might not override an "APPROVED" state in GitHub's view, but strictly speaking it's a new review.
			// Let's assume the latest review state is the current state for simplicity, checking specifically for key states.

			// Simplified logic: Just take the latest state.
			// Exception: If state is "COMMENTED", it might not potentially revoke "APPROVED".
			// But for this tool, if someone comments after approving, maybe they are just chatting.
			// Safety First: If we want to be safe, maybe we should only count APPROVED if it is the LATEST action?
			// Or should we ignore COMMENTED if there was a previous APPROVED?
			// A "review" with "COMMENT" state doesn't necessarily change the PR status to "Changes Requested".
			// Let's stick thereto: If Latest is APPROVED, they are approved.
			// If Latest is CHANGES_REQUESTED, they are blocking.
			// If Latest is COMMENTED... it depends.
			//
			// Improve logic:
			// If we have an existing state for author:
			//   If new state is APPROVED -> Update.
			//   If new state is CHANGES_REQUESTED -> Update.
			//   If new state is DISMISSED -> Update (effectively clears previous).
			//   If new state is COMMENTED -> Keep previous status if it was APPROVED or CHANGES_REQUESTED.
			//      (Comments usually don't change the voting status).

			currentState, exists := reviewsByAuthor[user]
			if !exists {
				reviewsByAuthor[user] = state
				continue
			}

			// If already exists, decide whether to overwrite
			if state == "COMMENTED" {
				// Don't overwrite APPROVED or CHANGES_REQUESTED with COMMENTED
				if currentState != "APPROVED" && currentState != "CHANGES_REQUESTED" {
					reviewsByAuthor[user] = state
				}
			} else {
				// Overwrite for APPROVED, CHANGES_REQUESTED, DISMISSED, etc.
				reviewsByAuthor[user] = state
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return reviewsByAuthor, nil
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
