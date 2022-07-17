package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
)

func main() {
	s := NewSender()
	if err := s.SendEmail(); err != nil {
		log.Fatal(err)
	}
}

type Sender interface {
	SendEmail() error
}
type mailCatcher struct{}

func NewSender() Sender {
	return &mailCatcher{}
}

func (mc *mailCatcher) SendEmail() error {

	ev := NewEnvelope("hoge@example.com", []string{"foo@example.com"}, "test subject", "tls test mail")
	smtpConfig := NewSmtpConfig(os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASSWORD"), "postfix", 587)

	auth := smtp.PlainAuth("", smtpConfig.User(), smtpConfig.Password(), smtpConfig.Host())

	c, err := smtp.Dial(smtpConfig.Addr())
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpConfig.Host(),
	}
	c.StartTLS(tlsConfig)

	if err = c.Auth(auth); err != nil {
		return err
	}
	if err = c.Mail(ev.From()); err != nil {
		return err
	}
	if err = c.Rcpt(ev.To()[0]); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(ev.Message()))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()
	log.Printf("mail send succeeded")

	return nil
}

type SmtpConfig struct {
	user     string
	password string
	host     string
	port     int
	addr     string
}

func NewSmtpConfig(user, passoword, host string, port int) *SmtpConfig {

	addr := fmt.Sprintf("%s:%d", host, port)
	return &SmtpConfig{
		user:     user,
		password: passoword,
		host:     host,
		port:     port,
		addr:     addr,
	}
}

func (sc *SmtpConfig) User() string     { return sc.user }
func (sc *SmtpConfig) Password() string { return sc.password }
func (sc *SmtpConfig) Host() string     { return sc.host }
func (sc *SmtpConfig) Port() int        { return sc.port }
func (sc *SmtpConfig) Addr() string     { return sc.addr }

type Envelope struct {
	from    string
	to      []string
	subject string
	body    string
	message string
}

func NewEnvelope(from string, to []string, subject string, body string) *Envelope {

	msg := fmt.Sprintf(`
From: %s
To: %s
%s
%s
`, from, to, subject, body)
	return &Envelope{
		from:    from,
		to:      to,
		subject: subject,
		body:    body,
		message: msg,
	}
}

func (e *Envelope) From() string    { return e.from }
func (e *Envelope) To() []string    { return e.to }
func (e *Envelope) Subject() string { return e.subject }
func (e *Envelope) Message() []byte { return []byte(e.message) }
