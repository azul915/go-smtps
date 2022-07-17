package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"os"
)

func main() {

	e := NewEnvelope("hoge@example.com", []string{"foo@example.com", "bar@example.com"}, "test subject", "tls test mail")
	c := NewSmtpConfig(os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASSWORD"), "postfix", 587)
	s := NewSender(e, c)
	if err := s.SendEmail(); err != nil {
		log.Fatal(err)
	}
}

type Sender interface {
	SendEmail() error
}
type mailDev struct {
	Envelope
	SmtpConfig
}

func NewSender(envelope *Envelope, smtpConfig *SmtpConfig) Sender {
	return &mailDev{
		Envelope:   *envelope,
		SmtpConfig: *smtpConfig,
	}
}

func (mc *mailDev) SendEmail() error {

	c, err := smtp.Dial(string(mc.SmtpConfig.Addr()))
	if err != nil {
		return err
	}
	defer c.Close()

	// EHLO
	if err = c.Hello(string(mc.SmtpConfig.Host())); err != nil {
		return err
	}
	// STARTTLS
	if err = c.StartTLS(&tls.Config{
		InsecureSkipVerify: true,
		ServerName:         string(mc.SmtpConfig.Host()),
	}); err != nil {
		return err
	}

	if _, ok := c.TLSConnectionState(); ok {
		log.Println("with SMTP over SSL/TLS(TLS 1.2)")
	} else {
		log.Println("with SMTP (plain text)")
	}

	// AUTH PLAIN
	auth := smtp.PlainAuth("", string(mc.SmtpConfig.User()), string(mc.SmtpConfig.Password()), string(mc.SmtpConfig.Host()))
	if err = c.Auth(auth); err != nil {
		return err
	}

	for _, v := range mc.Envelope.To() {
		// RSET
		if err = c.Reset(); err != nil {
			return err
		}
		// MAIL
		log.Printf("From: %v\n", mc.Envelope.From())
		if err = c.Mail(string(mc.Envelope.From())); err != nil {
			return err
		}
		// RCPT
		log.Printf("To: %v\n", v.Address)
		if err = c.Rcpt(v.Address); err != nil {
			return err
		}
		// DATA
		w, err := c.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(mc.Envelope.Message())
		if err != nil {
			return err
		}
		w.Close()
	}

	// QUIT
	if err = c.Quit(); err != nil {
		return err
	}
	log.Printf("mail send succeeded")
	return nil
}

type (
	User     string
	Password string
	Host     string
	Port     int
	Address  string
)

type SmtpConfig struct {
	user     User
	password Password
	host     Host
	port     Port
	addr     Address
}

func NewSmtpConfig(user, password, host string, port int) *SmtpConfig {

	addr := fmt.Sprintf("%s:%d", host, port)
	return &SmtpConfig{
		user:     User(user),
		password: Password(password),
		host:     Host(host),
		port:     Port(port),
		addr:     Address(addr),
	}
}

func (sc *SmtpConfig) User() User         { return sc.user }
func (sc *SmtpConfig) Password() Password { return sc.password }
func (sc *SmtpConfig) Host() Host         { return sc.host }
func (sc *SmtpConfig) Port() Port         { return sc.port }
func (sc *SmtpConfig) Addr() Address      { return sc.addr }

type (
	From    string
	To      []mail.Address
	Subject string
	Body    string
	Message []byte
)

type Envelope struct {
	from    From
	to      To
	subject Subject
	body    Body
}

func NewEnvelope(from string, to []string, subject string, body string) *Envelope {

	mas := make([]mail.Address, len(to))
	for i := range to {
		mas[i] = mail.Address{Name: "", Address: to[i]}
	}
	return &Envelope{
		from:    From(from),
		to:      mas,
		subject: Subject(subject),
		body:    Body(body),
	}
}

func (e *Envelope) From() From       { return e.from }
func (e *Envelope) To() To           { return e.to }
func (e *Envelope) Subject() Subject { return e.subject }
func (e *Envelope) Message() Message {
	msg := fmt.Sprintf("Subject: %s\r\n\r\n%s\r\n", e.subject, e.body)
	return []byte(msg)
}
