package api

// import (
// 	"bytes"
// 	"errors"
// 	"html/template"

// 	log "github.com/sirupsen/logrus"
// 	"github.com/hectane/go-nonblockingchan"
// 	"github.com/sendgrid/sendgrid-go"
// 	"github.com/j0ni/service-skeleton-go/olio/util"
// )

// type EmailAPI struct {
// 	sendGridKey        string
// 	outgoingMail       *nbc.NonBlockingChan
// 	templates          map[string]*template.Template
// 	SenderFunction     func(*Email) error
// }

// type Email struct {
// 	IsHtml         bool
// 	Subject        string
// 	Body           string
// 	Recipients     []string
// 	From           string
// 	FromName       string
// 	TemplateID     string
// 	TemplateValues map[string]string
// }

// ////////////////////////////////////////////////////////////////////////////////////////
// // public
// ////////////////////////////////////////////////////////////////////////////////////////

// //
// // Constructor
// //
// func NewEmailAPI(templates []string) *EmailAPI {
// 	api := EmailAPI{}
// 	api.outgoingMail = nbc.New()
// 	api.templates = make(map[string]*template.Template)

// 	if util.GetEnv("GIN_MODE", "") == "test" {
// 		return &api
// 	}

// 	// initialize templates
// 	for _, template := range templates {
// 		api.loadTemplate(template)
// 	}

// 	sendGridKey := util.GetEnv("SENDGRID_API_KEY", "")
// 	if sendGridKey == "" {
// 		log.Warn("No SENGRID_API_KEY defined; no emails will be sent.")
// 	} else {
// 		api.sendGridKey = sendGridKey
// 		go api.listenForEmails()
// 	}

// 	api.SenderFunction = api.doSend

// 	return &api
// }

// func (self *EmailAPI) EmailEnabled() bool {
// 	return self.sendGridKey != ""
// }

// func (self *EmailAPI) listenForEmails() {
// 	// forever listen
// 	for true {
// 		v, ok := <-self.outgoingMail.Recv
// 		if !ok {
// 			continue
// 		}

// 		email := v.(*Email)

// 		err := self.SenderFunction(email)
// 		if err != nil {
// 			log.Error("Failed to send email", err)
// 		}
// 	}
// }

// func (self *EmailAPI) loadTemplate(name string) (*template.Template, error) {
// 	webResourcesDir := util.GetEnv("WEB_RESOURCES_PATH", "")
// 	path := webResourcesDir + "/templates/email/" + name + ".html"
// 	log.Info("Loading email template " + path)
// 	t, err := template.ParseFiles(path)
// 	if err == nil && t != nil {
// 		self.templates[name] = t
// 	} else {
// 		log.Warn("Could not find template " + name)
// 		return nil, err
// 	}

// 	return t, nil
// }

// func (self *EmailAPI) SendSimpleMail(isHtml bool, subject string, body string, recipients []string, from string, fromName string) error {
// 	email := Email{}
// 	email.From = from
// 	email.FromName = fromName
// 	email.Recipients = recipients
// 	email.Body = body
// 	email.Subject = subject
// 	email.IsHtml = isHtml

// 	// send the email along to be sent asynchronously
// 	self.outgoingMail.Send <- &email

// 	return nil
// }

// func (self *EmailAPI) SendTemplatedEmail(isHtml bool, subject string, templateId string, templateValues map[string]string, recipients []string, from string, fromName string, async bool) error {
// 	email := Email{}
// 	email.From = from
// 	email.FromName = fromName
// 	email.Recipients = recipients
// 	email.TemplateID = templateId
// 	email.TemplateValues = templateValues
// 	email.Subject = subject
// 	email.IsHtml = isHtml

// 	if async {
// 		// send the email along to be sent asynchronously
// 		self.outgoingMail.Send <- &email
// 	} else {
// 		return self.SenderFunction(&email)
// 	}

// 	return nil
// }

// func (self *EmailAPI) doSend(email *Email) error {
// 	recipients := email.Recipients
// 	subject := email.Subject
// 	body := email.Body
// 	isHtml := email.IsHtml
// 	from := email.From
// 	templateId := email.TemplateID
// 	templateValues := email.TemplateValues

// 	message := sendgrid.NewMail()

// 	for _, recipient := range recipients {
// 		message.AddTo(recipient) // Returns error if email string is not valid RFC 5322
// 	}

// 	message.SetSubject(subject)

// 	if templateId != "" {
// 		// parse template
// 		template := self.templates[templateId]
// 		if template == nil {
// 			return errors.New("No such template " + templateId + "; cannot send email.")
// 		}

// 		writer := new(bytes.Buffer)
// 		template.Execute(writer, templateValues)

// 		body = string(writer.Bytes())

// 		log.Debug(body)
// 	}

// 	if isHtml {
// 		message.SetHTML(body)
// 	} else {
// 		message.SetText(body)
// 	}

// 	message.SetFrom(from)
// 	if email.FromName != "" {
// 		message.SetFromName(email.FromName)
// 	}

// 	if !self.EmailEnabled() {
// 		return errors.New("Email not enabled")
// 	}

// 	sg := sendgrid.NewSendGridClientWithApiKey(self.sendGridKey)
// 	if r := sg.Send(message); r != nil {
// 		log.Error("Email sent error:", r)
// 		return r
// 	}

// 	log.Debug("Email sent.")

// 	return nil
// }
