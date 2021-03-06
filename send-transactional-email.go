package sendmail

import (
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"encoding/base64"
	"github.com/sendgrid/sendgrid-go"
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
	ReplyToFullName   string
	ReplyToEmail      string
	RecipientFullName string
	RecipientEmail    string
	Subject           string
	TransactionalId   string
	Content           string
	Personalization   []Personalization
	Attachments       []FileInfo
}

type FileInfo struct {
	/*
	* ||| How to get a File []byte: |||
	*
	* -> FOR MULTIPART FILES
	*
	* buf := bytes.NewBuffer(nil)
	* io.Copy(buf, MultipartFile)
	* FileInfo.File = buf.Bytes()
	*
	* -> FOR os.FILES
	*
	* FileInfo.File, err := ioutil.ReadFile(fileOnDiskName)
	*/
	File []byte
	Name string
	Type FileType
}

type Personalization struct {
	Name  string
	Value string
}

// SendTransactional send transactional email using Sendgrid
func SendTransactional(sendgridKey string, emailInfo TransactionalEmailInfo) (*rest.Response, error) {

	// Create email
	m := mail.NewV3Mail()
	m.SetFrom(mail.NewEmail(emailInfo.SenderFullName, emailInfo.SenderEmail))
	if emailInfo.ReplyToEmail != "" {
		if emailInfo.ReplyToFullName != "" {
			m.SetReplyTo(mail.NewEmail(emailInfo.ReplyToFullName, emailInfo.RecipientEmail))
		} else {
			return nil, errors.New("Missing replier full name.")
		}
	}
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
		attachment := mail.NewAttachment()
		if emailInfo.Attachments[k].Type == PDF {
			encoded := base64.StdEncoding.EncodeToString([]byte(emailInfo.Attachments[k].File))
			attachment.SetContent(encoded)
			attachment.SetType("application/pdf")
			attachment.SetFilename(emailInfo.Attachments[k].Name + ".pdf")
			attachment.SetDisposition("attachment")
			attachment.SetContentID("Attachment")
		} else {
			return nil, errors.New("Unknown file type '" + emailInfo.Attachments[k].Type + "'")
		}
		m.Attachments = append(m.Attachments, attachment)
	}

	request := sendgrid.GetRequest(sendgridKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	return response, err
}
