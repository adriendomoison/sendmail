package sendmail

import (
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"encoding/base64"
	"github.com/sendgrid/sendgrid-go"
	"bytes"
	"io"
	"mime/multipart"
	"github.com/go-errors/errors"
	"github.com/sendgrid/rest"
)

type FileType string

const (
	PDF FileType = "PDF"
)

type TransactionalEmailInfo struct {
	SenderFullName    string
	SenderEmail       string
	RecipientFullName string
	RecipientEmail    string
	Subject           string
	TransactionalId   string
	Content           string
	Personalization   []Personalization
	Attachments       []FileInfo
}

type FileInfo struct {
	File multipart.File
	Name string
	Type FileType
}

type Personalization struct {
	Name  string
	Value string
}

func SendTransactional(sendgridKey string, emailInfo TransactionalEmailInfo) (*rest.Response, error) {

	// Create email
	m := mail.NewV3Mail()
	m.SetFrom(mail.NewEmail(emailInfo.SenderFullName, emailInfo.SenderEmail))
	if emailInfo.Subject != "" {
		m.Subject = emailInfo.Subject
	}
	p := mail.NewPersonalization()
	p.AddTos(mail.NewEmail(emailInfo.RecipientFullName, emailInfo.RecipientEmail))
	m.AddPersonalizations(p)
	if emailInfo.Content != "" {
		m.AddContent(mail.NewContent("text/html", emailInfo.Content))
	}

	// Filling personalization content
	for key := range emailInfo.Personalization {
		m.Personalizations[0].SetSubstitution(emailInfo.Personalization[key].Name, emailInfo.Personalization[key].Value)
	}

	// Set template to use
	m.SetTemplateID(emailInfo.TransactionalId)

	//Attachments
	for k := range emailInfo.Attachments {
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, emailInfo.Attachments[k].File); err != nil {
			return nil, errors.New("File was corrupted")
		}

		if emailInfo.Attachments[k].Type == PDF {
			pdf := mail.NewAttachment()
			encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
			pdf.SetContent(encoded)
			pdf.SetType("application/pdf")
			pdf.SetFilename(emailInfo.Attachments[k].Name + ".pdf")
			pdf.SetDisposition("attachment")
			pdf.SetContentID("Attachment")
		} else {
			return nil, errors.New("Unknown file type '" + emailInfo.Attachments[k].Type + "'")
		}
	}

	request := sendgrid.GetRequest(sendgridKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	return response, err
}
