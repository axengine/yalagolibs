package gmail

import (
	"gopkg.in/gomail.v2"
)

type GMail struct {
	smtp     string
	port     int
	sender   string
	password string
}

func New(smtp string, port int, sender, password string) *GMail {
	return &GMail{
		smtp:     smtp,
		port:     port,
		sender:   sender,
		password: password,
	}
}

func (gmail *GMail) SendMail(from string, tos []string, subject string, msg string) error {
	m := gomail.NewMessage()
	if from == "" {
		from = gmail.sender
	}
	m.SetHeader("From", from)
	for _, to := range tos {
		m.SetHeader("To", to)
	}
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", msg)
	d := gomail.NewDialer(gmail.smtp, gmail.port, gmail.sender, gmail.password)
	return d.DialAndSend(m)
}
