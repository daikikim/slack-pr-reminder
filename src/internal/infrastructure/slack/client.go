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
		log.Printf("[SLACK] [DRY-RUN] Would send to channel %s, mentioning %s", c.channel, slackID)
		log.Printf("[SLACK] [DRY-RUN] Message: %s", message)
		return nil
	}

	log.Printf("[SLACK] Sending reminder to channel: %s, mentioning user: %s", c.channel, slackID)
	log.Printf("[SLACK] Message length: %d characters", len(message))
	_, _, err := c.client.PostMessageContext(
		ctx,
		c.channel,
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		log.Printf("[SLACK] Error sending message: %v", err)
		return err
	}

	log.Printf("[SLACK] Successfully sent reminder to channel %s, mentioning %s", c.channel, slackID)
	return nil
}
