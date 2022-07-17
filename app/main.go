package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
)

func main() {

	e := NewEnvelope("hoge@example.com", []string{"foo@example.com"}, "test subject", "tls test mail")

	// c, err := NewSmtpConfig(os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASSWORD"), os.Getenv("SMTP_DOMAIN"), os.Getenv("SMTP_PORT"))
	c, err := NewSmtpConfig(os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASSWORD"), "postfix", "587")
	if err != nil {
		log.Fatal(err)
	}

	s := NewSender(e, c)
	if err = s.SendEmail(); err != nil {
		log.Fatal(err)
	}
}

type Sender interface {
	Envelope() *envelope
	Config() *smtpConfig
	SendEmail() error
}
type mailDev struct {
	envelope
	smtpConfig
}

func NewSender(envelope *envelope, smtpConfig *smtpConfig) Sender {
	return &mailDev{
		envelope:   *envelope,
		smtpConfig: *smtpConfig,
	}
}

func (mc *mailDev) Envelope() *envelope { return &mc.envelope }
func (mc *mailDev) Config() *smtpConfig { return &mc.smtpConfig }
func (mc *mailDev) SendEmail() error {

	smtpConfig, evlp := mc.Config(), mc.Envelope()

	c, err := smtp.Dial(string(smtpConfig.Addr()))
	if err != nil {
		return err
	}
	defer c.Close()

	// EHLO
	if err = c.Hello(string(smtpConfig.Host())); err != nil {
		return err
	}
	// STARTTLS
	if err = c.StartTLS(&tls.Config{
		// TODO: certification
		InsecureSkipVerify: true,
		ServerName:         string(smtpConfig.Host()),
	}); err != nil {
		return err
	}

	if _, ok := c.TLSConnectionState(); ok {
		log.Println("with SMTP over SSL/TLS(TLS 1.2)")
	} else {
		log.Println("with SMTP (plain text)")
	}

	// AUTH PLAIN
	// auth := smtp.PlainAuth("", string(mc.User()), string(mc.Password()), string(mc.Host()))
	// if err = c.Auth(auth); err != nil {
	// 	return err
	// }
	// AUTH CRAM-MD5
	auth := smtp.CRAMMD5Auth(string(smtpConfig.User()), string(smtpConfig.Password()))
	if err = c.Auth(auth); err != nil {
		return err
	}

	// toAsString := make([]string, len(mc.Envelope.To()))
	// for i := range mc.Envelope.To() {
	// 	toAsString[i] = string(mc.Envelope.To()[i].Address)
	// }
	// if err = smtp.SendMail(string(mc.SmtpConfig.Addr()), auth, string(mc.Envelope.From()), toAsString, mc.Envelope.Message()); err != nil {
	// 	return err
	// }

	for _, v := range evlp.To() {
		// RSET
		if err = c.Reset(); err != nil {
			return err
		}
		// MAIL
		log.Printf("From: %v\n", evlp.From())
		if err = c.Mail(string(evlp.From())); err != nil {
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
		_, err = w.Write(evlp.Message())
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

type smtpConfig struct {
	user     User
	password Password
	host     Host
	port     Port
	addr     Address
}

func NewSmtpConfig(user, password, host string, port string) (*smtpConfig, error) {

	p, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", host, p)
	return &smtpConfig{
		user:     User(user),
		password: Password(password),
		host:     Host(host),
		port:     Port(p),
		addr:     Address(addr),
	}, nil
}

func (sc *smtpConfig) User() User         { return sc.user }
func (sc *smtpConfig) Password() Password { return sc.password }
func (sc *smtpConfig) Host() Host         { return sc.host }
func (sc *smtpConfig) Port() Port         { return sc.port }
func (sc *smtpConfig) Addr() Address      { return sc.addr }

type (
	From    string
	To      []mail.Address
	Subject string
	Body    string
	Message []byte
)

type envelope struct {
	from    From
	to      To
	subject Subject
	body    Body
}

func NewEnvelope(from string, to []string, subject string, body string) *envelope {

	mas := make([]mail.Address, len(to))
	for i := range to {
		mas[i] = mail.Address{Name: "", Address: to[i]}
	}
	return &envelope{
		from:    From(from),
		to:      mas,
		subject: Subject(subject),
		body:    Body(body),
	}
}

func (e *envelope) From() From       { return e.from }
func (e *envelope) To() To           { return e.to }
func (e *envelope) Subject() Subject { return e.subject }
func (e *envelope) Message() Message {

	msg := bytes.NewBuffer([]byte(""))
	msg.WriteString(fmt.Sprintf("From: %s\r\n", e.from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", e.to))
	msg.WriteString(fmt.Sprintf("Bcc: %s\r\n", ""))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", e.subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: base64\r\n")
	msg.WriteString("\r\n")
	return msg.Bytes()
}
