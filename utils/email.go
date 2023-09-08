package utils

import (
	"crypto/tls"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/jordan-wright/email"
	"net/smtp"
	"net/textproto"
)

type EmailData struct {
	SendTo  string
	Subject string
	Tittle  string
	Content string
}

func SendEmail(cfg config.EmailConfig, data EmailData) error {
	message := &email.Email{
		To:      []string{data.SendTo},
		From:    fmt.Sprintf("%s <%s>", cfg.Name, cfg.Username),
		Subject: data.Subject,
		Text:    []byte(data.Tittle),
		HTML:    []byte(data.Content),
		Headers: textproto.MIMEHeader{},
	}

	// smtp.PlainAuth：the first param can be empty，the second param should be the email account，the third param is the secret of the email
	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)

	return message.SendWithTLS(addr, auth, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         cfg.SMTPHost,
	})
}
