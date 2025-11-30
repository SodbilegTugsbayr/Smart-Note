package mailer

import (
	"log"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func (mailer *Mailer) Send(to, subject, message string) bool {
	m := gomail.NewMessage()
	m.SetHeader("From", mailer.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", message)

	d := gomail.NewDialer(mailer.Host, mailer.Port, mailer.Username, mailer.Password)
	// Send emails using d.

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Error at sending message (%s) to '%s'. Error:%s.", message, to, err)
		return false
	}
	log.Printf("Message to %s.", to)
	return true
}
