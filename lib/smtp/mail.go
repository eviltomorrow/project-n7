package smtp

import (
	"crypto/tls"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"gopkg.in/gomail.v2"
)

const (
	TextHTML  = "text/html"
	TextPlain = "text/plain"
)

// SendWithoutSSL send mail
func SendWithoutSSL(smtpAddress string, username, password string, message *Message) error {
	return send(smtpAddress, 25, username, password, false, message)
}

// SendWithoutSSL send mail
func SendWithSSL(smtpAddress string, username, password string, message *Message) error {
	return send(smtpAddress, 587, username, password, true, message)
}

func send(smtpAddress string, port int, username, password string, ssl bool, message *Message) error {
	if err := verify(message); err != nil {
		return err
	}
	var m = gomail.NewMessage()
	setContact(m, "From", message.From)
	setContact(m, "To", message.To...)
	setContact(m, "Cc", message.Cc...)
	setContact(m, "Bcc", message.Bcc...)
	m.SetHeader("Subject", message.Subject)
	m.SetBody(message.ContentType, message.Body)
	setFile(m, message.Attach)

	var dialer = gomail.Dialer{
		Host:     smtpAddress,
		Port:     port,
		Username: username,
		Password: password,
		// SSL:      ssl,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return dialer.DialAndSend(m)
}

func setFile(m *gomail.Message, files []File) {
	if len(files) == 0 {
		return
	}
	for _, file := range files {
		if file.Path == "" {
			continue
		}
		if file.AliasName == "" {
			m.Attach(file.Path)
		} else {
			m.Attach(file.Path, gomail.Rename(file.AliasName))
		}
	}
}

func setContact(m *gomail.Message, field string, contacts ...Contact) {
	if len(contacts) == 0 {
		return
	}
	var values = make([]string, 0, len(contacts))
	for _, contact := range contacts {
		if contact.Address == "" {
			continue
		}
		if contact.Name == "" {
			values = append(values, contact.Address)
		} else {
			values = append(values, m.FormatAddress(contact.Address, contact.Name))
		}
	}
	if len(values) == 0 {
		return
	}
	m.SetHeader(field, values...)
}

func verify(message *Message) error {
	if message.From.Address == "" {
		return fmt.Errorf("from address is missing")
	}

	var hasTo bool
	for _, to := range message.To {
		if to.Address != "" {
			hasTo = true
		}
	}
	if !hasTo {
		return fmt.Errorf("to address is missing")
	}

	switch message.ContentType {
	case TextHTML, TextPlain:
	default:
		return fmt.Errorf("ContentType is invalid")
	}
	return nil
}

// Message msg
type Message struct {
	From        Contact   `json:"from"`
	To          []Contact `json:"to"`
	Cc          []Contact `json:"cc"`
	Bcc         []Contact `json:"bcc"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	ContentType string    `json:"content-type"`
	Attach      []File    `json:"attach"`
}

func (m *Message) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(m)
	return string(buf)
}

// Contact contact
type Contact struct {
	Name    string `json:"name"`
	Address string `json:"access"`
}

// File file
type File struct {
	Path      string `json:"path"`
	AliasName string `json:"alias-name"`
}
