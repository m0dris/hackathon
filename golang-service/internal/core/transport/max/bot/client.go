package core_max_bot

import (
	"context"
	"fmt"
	"net/http"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

var updateTypes = []string{
	string(schemes.TypeBotStarted),
	string(schemes.TypeMessageCreated),
	string(schemes.TypeMessageCallback),
}

type Client struct {
	api    *maxbot.Api
	config Config
}

func NewClient(config Config) (*Client, error) {
	api, err := maxbot.New(config.Token)
	if err != nil {
		return nil, fmt.Errorf("init MAX bot api client: %w", err)
	}

	return &Client{api: api, config: config}, nil
}

func (c *Client) WebhookHandler(handle func(update schemes.UpdateInterface)) http.HandlerFunc {
	return c.api.GetUpdateHandlerFunc(handle, c.config.Secret)
}

func (c *Client) Subscribe(ctx context.Context) error {
	_, err := c.api.Subscriptions.Subscribe(ctx, c.config.WebhookURL, updateTypes, c.config.Secret)
	if err != nil {
		return fmt.Errorf("subscribe webhook: %w", err)
	}

	return nil
}

func (c *Client) SendText(ctx context.Context, chatID int64, text string) error {
	m := maxbot.NewMessage().SetChat(chatID).SetText(text)

	if err := c.api.Messages.Send(ctx, m); err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (c *Client) SendKeyboard(ctx context.Context, chatID int64, text string, keyboard *maxbot.Keyboard) error {
	m := maxbot.NewMessage().SetChat(chatID).SetText(text).AddKeyboard(keyboard)

	if err := c.api.Messages.Send(ctx, m); err != nil {
		return fmt.Errorf("send message with keyboard: %w", err)
	}

	return nil
}

func (c *Client) AnswerCallback(ctx context.Context, callbackID string, notification string) error {
	_, err := c.api.Messages.AnswerOnCallback(ctx, callbackID, &schemes.CallbackAnswer{
		Notification: notification,
	})
	if err != nil {
		return fmt.Errorf("answer callback: %w", err)
	}

	return nil
}
