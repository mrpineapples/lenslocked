package email

import (
	"context"
	"fmt"
	"time"

	"github.com/mailgun/mailgun-go/v3"
)

const (
	welcomeSubject = "Welcome to lens-locked.com!"
)

const welcomeText = `Hi There!

Welcome to lens-locked.com! We really hope you enjoy using
our application!

Best,
Michael
`

const welcomeHTML = `Hi there! <br/>
<br/>
Welcome to
<a href="https://lens-locked.com">lens-locked.com</a>! We really hope you enjoy using our application!<br/>
<br/>
Best,<br/>
Michael
`

func WithMailgun(domain, apiKey, publicKey string) ClientConfig {
	return func(c *Client) {
		mg := mailgun.NewMailgun(domain, apiKey)
		c.mg = mg
	}
}

func WithSender(name, email string) ClientConfig {
	return func(c *Client) {
		c.from = buildEmail(name, email)
	}
}

type ClientConfig func(*Client)

func NewClient(opts ...ClientConfig) *Client {
	client := Client{
		from: "support@lens-locked.com",
	}
	for _, opt := range opts {
		opt(&client)
	}

	return &client
}

type Client struct {
	from string
	mg   mailgun.Mailgun
}

func (c *Client) Welcome(toName, toEmail string) error {
	message := c.mg.NewMessage(c.from, welcomeSubject, welcomeText, buildEmail(toName, toEmail))
	message.SetHtml(welcomeHTML)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, _, err := c.mg.Send(ctx, message)
	return err
}

func buildEmail(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}
