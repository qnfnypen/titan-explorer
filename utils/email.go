package utils

import (
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
	e := &email.Email{
		To:      []string{data.SendTo},
		From:    cfg.Name + "<" + cfg.Address + ">",
		Subject: data.Subject,
		Text:    []byte(data.Tittle),
		HTML:    []byte("<h1>" + data.Content + "</h1>\n"),
		Headers: textproto.MIMEHeader{},
	}

	// send function：
	// smtp.PlainAuth：the first param can be empty，the second param should be the email account，the third param is the secret of the email
	err := e.Send(cfg.SMTP, smtp.PlainAuth("", cfg.Address, cfg.Secret, cfg.Host))
	if err != nil {
		return err
	}
	return nil

}
