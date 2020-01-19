package email

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/mailgun/mailgun-go/v3"
)

const (
	welcomeSubject = "Welcome to lens-locked.com!"
	resetSubject   = "Instructions for resetting your password."
	resetBaseURL   = "https://lens-locked.com/reset"
)

const welcomeText = `Hi There!

Welcome to lens-locked.com! We really hope you enjoy using
our application!

Best,
lens-locked Support
`

const welcomeHTML = `Hi there! <br/>
<br/>
Welcome to
<a href="https://lens-locked.com">lens-locked.com</a>! We really hope you enjoy using our application!<br/>
<br/>
Best,<br/>
lens-locked Support
`

const resetTextTmpl = `Hi there!

It appears that you have requested a password reset. If this was you, please follow the link below to update your password:

%s

If you are asked for a token, please use the following value:

%s

If you didn't request a password reset you can safely ignore this email and your account with not be changes.

Best,
lens-locked Support
`
const resetHTMLTmpl = `Hi there!<br/>
<br/>
It appears that you have requested a password reset. If this was you, please follow the link below to update your password:<br/>
<br/>
<a href="%s">%s</a><br/>
<br/>
If you are asked for a token, please use the following value:<br/>
<br/>
%s<br/>
<br/>
If you didn't request a password reset you can safely ignore this email and your account with not be changed.<br/>
<br/>
Best,<br/>
lens-locked Support<br/>
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

func (c *Client) ResetPw(toEmail, token string) error {
	v := url.Values{}
	v.Set("token", token)
	resetURL := resetBaseURL + "?" + v.Encode()
	resetText := fmt.Sprintf(resetTextTmpl, resetURL, token)
	message := c.mg.NewMessage(c.from, resetSubject, resetText, toEmail)

	resetHTML := fmt.Sprintf(resetHTMLTmpl, resetURL, resetURL, token)
	message.SetHtml(resetHTML)

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
