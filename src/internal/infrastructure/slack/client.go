package slack

import (
	"context"
	"log"

	"github.com/slack-go/slack"
)

// Client implements model.Notifier.
type Client struct {
	client  *slack.Client
	channel string
	dryRun  bool
}

// NewClient creates a new Slack client.
func NewClient(token, channel string, dryRun bool) *Client {
	return &Client{
		client:  slack.New(token),
		channel: channel,
		dryRun:  dryRun,
	}
}

// SendReminder sends a reminder message to a user.
func (c *Client) SendReminder(ctx context.Context, slackID string, message string) error {
	if c.dryRun {
		log.Printf("[DRY-RUN] Would send to %s: %s", slackID, message)
		return nil
	}

	_, _, err := c.client.PostMessageContext(
		ctx,
		c.channel,
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		return err
	}

	log.Printf("Sent reminder to %s", slackID)
	return nil
}
