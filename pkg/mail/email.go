package mail

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailMessage 内容
type EmailMessage struct {
	From        string
	Nickname    string
	To          []string
	Cc          []string
	Subject     string
	ContentType string
	Content     string
	Attach      string
}

// NewEmailMessage 返回消息对象
// from: 发件人
// subject: 标题
// contentType: 内容的类型 text/plain text/html
// attach: 附件
// to: 收件人
// cc: 抄送人
func NewEmailMessage(from, nickname, subject, contentType, content, attach string, to, cc []string) *EmailMessage {
	return &EmailMessage{
		From:        from,
		Nickname:    nickname,
		Subject:     subject,
		ContentType: contentType,
		Content:     content,
		To:          to,
		Cc:          cc,
		Attach:      attach,
	}
}

// EmailClient 发送客户端
type EmailClient struct {
	Host     string
	Port     int
	Username string
	Password string
	Message  *EmailMessage
}

// NewEmailClient 返回一个邮件客户端
// host smtp地址
// username 用户名
// password 密码
// port 端口
func NewEmailClient(host, username, password string, port int, message *EmailMessage) *EmailClient {
	return &EmailClient{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Message:  message,
	}
}

// SendMessage 发送邮件
func (c *EmailClient) SendMessage() (bool, error) {
	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)
	cc := strings.Join(c.Message.Cc, ";")
	to := strings.Join(c.Message.To, ";")
	if c.Message.Nickname == "" {
		c.Message.Nickname = c.Message.From
	}
	from := fmt.Sprintf("%s <%s>", c.Message.Nickname, c.Message.From)
	msg := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + c.Message.Subject + "\r\n" +
		"Cc: " + cc + "\r\n" +
		"Content-Type: " + c.Message.ContentType + "; charset=UTF-8" + "\r\n" +
		"\r\n" +
		c.Message.Content + "\r\n",
	)

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	err := smtp.SendMail(addr, auth, c.Message.From, c.Message.To, msg)
	if err != nil {
		return false, err
	}

	return true, nil
}
