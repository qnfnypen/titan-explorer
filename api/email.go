package api

import (
	_ "embed"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/mail"
	"github.com/pkg/errors"
	"math/rand"
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

	var mailCfg config.EmailConfig
	if len(config.Cfg.Emails) > 0 {
		mailCfg = config.Cfg.Emails[rand.Intn(len(config.Cfg.Emails))]
	} else {
		log.Errorf("email config not set")
		return errors.Errorf("email config not set")
	}

	port, err := strconv.ParseInt(mailCfg.SMTPPort, 10, 64)
	if err != nil {
		log.Errorf("parse port: %v", err)
	}

	message := mail.NewEmailMessage(mailCfg.From, mailCfg.Nickname, emailSubject[lang], contentType, content, "", []string{sendTo}, nil)
	client := mail.NewEmailClient(mailCfg.SMTPHost, mailCfg.Username, mailCfg.Password, int(port), message)
	_, err = client.SendMessage()
	if err != nil {
		return err
	}

	return nil
}
