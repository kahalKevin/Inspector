package mailer

import (
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
)

//Mailer ...
type Mailer interface {
	Send(subject, body string) error
	Message(key string) (string, string)
}

//Pakpos data structure
type Pakpos struct {
	Address  string
	From     string
	To       []string
	ToHeader string
	Subject  map[string]string
	Body     map[string]string
}

//New ...
func New(host string, port int, from string, to []string) *Pakpos {
	var toAddress = make([]string, 0)

	fromAddress := mail.Address{
		Name:    from,
		Address: from,
	}

	for k := range to {
		m := mail.Address{
			Name:    to[k],
			Address: to[k],
		}
		toAddress = append(toAddress, m.String())
	}

	return &Pakpos{
		Address:  fmt.Sprintf("%s:%d", host, port),
		From:     fromAddress.String(),
		To:       toAddress,
		ToHeader: strings.Join(toAddress, ", "),
	}
}

//SetMessages ...
func (m *Pakpos) SetMessages(subject, body map[string]string) *Pakpos {
	m.Subject = subject
	m.Body = body
	return m
}

//Send email
func (m *Pakpos) Send(subject, body string) error {
	header := make(map[string]string)

	header["To"] = m.ToHeader
	header["From"] = m.From
	header["Subject"] = subject
	header["Content-Type"] = `text/html; charset="UTF-8"`

	msg := ""
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body
	bMsg := []byte(msg)

	c, err := smtp.Dial(m.Address)
	if err != nil {
		return err
	}
	defer c.Close()

	if err = c.Mail(m.From); err != nil {
		return err
	}

	for _, addr := range m.To {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(bMsg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	err = c.Quit()
	if err != nil {
		return err
	}

	return nil
}

//Message ...
func (m *Pakpos) Message(key string) (subject, body string) {
	if _, ok := m.Subject[key]; !ok {
		return
	}

	if _, ok := m.Body[key]; !ok {
		return
	}

	subject = m.Subject[key]
	body = m.Body[key]
	return
}
