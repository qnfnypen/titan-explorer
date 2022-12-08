package utils

import (
	"github.com/gnasnik/titan-explorer/config"
	"github.com/jordan-wright/email"
	"github.com/linguohua/titan/api"
	"net/smtp"
	"net/textproto"
)

type EmailData struct {
	SendTo  string
	Subject string
	Tittle  string
	Content string
}
type EmailConfig struct {
	Name    string
	Address string
	Secret  string
}

var EConfig EmailConfig

func Send(data EmailData) error {
	e := &email.Email{
		To:      []string{data.SendTo},
		From:    EConfig.Name + "<" + EConfig.Address + ">",
		Subject: data.Subject,
		Text:    []byte(data.Tittle),
		HTML:    []byte("your Device id and Secret：<h1>" + data.Content + "</h1>\n"),
		Headers: textproto.MIMEHeader{},
	}

	// send function：
	// smtp.qq.com:587：QQ email address and port
	// smtp.PlainAuth：the first param can be empty，the second param should be the email account，the third param is the secret of the email
	err := e.Send("smtp.qq.com:587", smtp.PlainAuth("", EConfig.Address, EConfig.Secret, "smtp.qq.com"))
	if err != nil {
		return err
	}
	return nil

}
func Demo() {
	emailSend := EmailData{"88486360@qq.com", "subject", "title", "others:you are so handsome"}
	Send(emailSend)
}

func EmailInit(cfg config.Config) {
	EConfig.Address = cfg.EmailAddress
	EConfig.Name = cfg.EmailName
	EConfig.Secret = cfg.EmailSecret
}

func HandleEmailInfo(sendTo string, results []api.NodeRegisterInfo) {
	var EData EmailData
	EData.Subject = "YOUR DEVICE INFO"
	EData.Tittle = "please check your device id and secret"
	EData.SendTo = sendTo
	EData.Content = ""
	for _, registration := range results {
		EData.Content += registration.DeviceID + ":" + registration.Secret + "\n"
	}
	err := Send(EData)
	if err != nil {
		log.Errorf("send email failed: %v", err)
		return
	}
}
