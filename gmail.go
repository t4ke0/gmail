package gmail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	gmailSMTPhost string = "smtp.gmail.com"
	gmailSMTPport string = "587"
)

const templateFormat string = `From: {{ .From }}
To: {{ .Recipient }}
Subject: {{ .Subject }}
MIME-Version: 1.0
Content-Type: multipart/mixed;charset=UTF-8;{{ .Boundary }}


--{{ .BoundaryRepr }}
Content-Type: text/plain
Content-Transfer-Encoding: 7bit

{{ .MessageText }}
`

// TemplateItems template fields to replace
type TemplateItems struct {
	EmailConfig
	Recipient string
}

// EmailConfig
type EmailConfig struct {
	From         string
	To           []string
	Subject      string
	MessageText  string
	Boundary     string
	BoundaryRepr string

	Attachements []string
}

// Email structure that represents an email. holds email creds and
// configuration
type Email struct {
	EmailConfig

	auth         smtp.Auth
	emailData    string
	processError error
}

// NewEmail construct a new email returns a pointer to Email structure.
func NewEmail(username, password string, cfg EmailConfig) *Email {
	auth := smtp.PlainAuth("", username, password, gmailSMTPhost)
	return &Email{
		EmailConfig: cfg,
		auth:        auth,
	}
}

// Marshal prepare email form
func (e *Email) Marshal() *Email {

	e.EmailConfig.BoundaryRepr = "MyBorder"
	e.EmailConfig.Boundary = fmt.Sprintf("boundary=\"%s\"", e.EmailConfig.BoundaryRepr)

	tpl := TemplateItems{
		EmailConfig: e.EmailConfig,
		Recipient:   strings.Join(e.To, ","),
	}

	tmpl, err := template.New("email").Parse(templateFormat)
	if err != nil {
		e.processError = err
		return e
	}

	buffer := &bytes.Buffer{}

	var hasAttachements bool

	if len(e.EmailConfig.Attachements) != 0 {
		hasAttachements = true
	}

	if err := tmpl.Execute(buffer, tpl); err != nil {
		e.processError = err
		return e
	}

	bufferData := buffer.String()
	if hasAttachements {
		for _, f := range e.EmailConfig.Attachements {
			data, err := os.ReadFile(f)
			if err != nil {
				e.processError = err
				return e
			}
			contentType := http.DetectContentType(data)
			b64Data := base64.StdEncoding.EncodeToString(data)

			bufferData += fmt.Sprintf(`--%s
Content-Type: %s
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename=%s

%s

`, e.EmailConfig.BoundaryRepr, contentType, filepath.Base(f), b64Data)
		}
	}

	bufferData += fmt.Sprintf("--%s--", e.EmailConfig.BoundaryRepr)
	e.emailData = bufferData

	return e
}

// Send sends the email.
func (e *Email) Send() *Email {
	addr := fmt.Sprintf("%s:%s", gmailSMTPhost, gmailSMTPport)
	if err := smtp.SendMail(addr,
		e.auth, e.EmailConfig.From,
		e.EmailConfig.To, []byte(e.emailData)); err != nil {
		e.processError = err
	}
	return e
}

// Error check if there is any error that occurs while processing the email or
// while sending it.
func (e *Email) Error() error {
	return e.processError
}
