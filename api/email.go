package api

import (
	_ "embed"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/mail"
	"strconv"
)

//go:embed template/en/mail.html
var contentEn string

//go:embed template/cn/mail.html
var contentCn string

func sendEmail(sendTo string, vc, lang string) error {
	emailSubject := map[string]string{
		"":               "[Titan Network] Your verification code",
		model.LanguageEN: "[Titan Network] Your verification code",
		model.LanguageCN: "[Titan Network] 您的验证码",
	}

	content := contentEn
	if lang == model.LanguageCN {
		content = contentCn
	}

	var verificationBtn = ""
	for _, code := range vc {
		verificationBtn += fmt.Sprintf(`<button class="button" th>%s</button>`, string(code))
	}
	content = fmt.Sprintf(content, verificationBtn)

	contentType := "text/html"
	port, err := strconv.ParseInt(config.Cfg.Email.SMTPPort, 10, 64)
	if err != nil {
		log.Errorf("parse port: %v", err)
	}
	message := mail.NewEmailMessage(config.Cfg.Email.From, config.Cfg.Email.Nickname, emailSubject[lang], contentType, content, "", []string{sendTo}, nil)
	client := mail.NewEmailClient(config.Cfg.Email.SMTPHost, config.Cfg.Email.Username, config.Cfg.Email.Password, int(port), message)
	_, err = client.SendMessage()
	if err != nil {
		return err
	}

	return nil
}
