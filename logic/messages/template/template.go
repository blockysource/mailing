package emailtemplate

import (
	"bytes"
	"net/http"
	"net/mail"
	"sync"
	"text/template"

	"github.com/sirupsen/logrus"

	mailprovider2 "github.com/blockysource/blocky/open-source/libs/blocky-cloud/email/message"

	"github.com/blockysource/blocky/services/mailing/public/mailingpb"
	"github.com/blockysource/mailing/logic/mailprovider"
)

// TemplateDefinition is an email template definition.
type TemplateDefinition struct {
	UID         string
	Name        string
	FromAddress string
	Subject     string
	Body        string
	Parameters  []Parameter
}

// TemplateParser is an email template.
type TemplateParser struct {
	l                   sync.RWMutex
	base                TemplateDefinition
	providerFromAddress *mail.Address
	fat                 *template.Template
	st                  *template.Template
	bt                  *template.Template
	log                 *logrus.Entry
	parameters          []Parameter
}

// ProviderUpdated updates the email template with the given provider.
func (t *TemplateParser) ProviderUpdated(p mailprovider.Provider) error {
	t.l.Lock()
	defer t.l.Unlock()

	t.providerFromAddress = p.GetDefaultFromAddress()
	return nil
}

// Parameter is a parameter of an email template.
type Parameter struct {
	Name         string
	DefaultValue string
}

var bytesBufferPool = sync.Pool{}

func getBuffer() *bytes.Buffer {
	b := bytesBufferPool.Get()
	if b == nil {
		return bytes.NewBuffer(nil)
	}
	buf := b.(*bytes.Buffer)
	return buf
}

func putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bytesBufferPool.Put(buf)
}

// Parse parses the email template and returns an email message.
func (t *TemplateParser) Parse(in *mailingpb.EnqueuedEmailMessage) (mailprovider2.Message, error) {
	// Try to parse all the parts of the template, starting with the FromAddress.
	buf := getBuffer()
	defer putBuffer(buf)

	// Parse the parameters.
	params := map[string]string{}
	for _, p := range t.parameters {
		params[p.Name] = p.DefaultValue
	}
	for _, p := range in.Parameters {
		params[p.Name] = p.Value
	}

	fromAddr := t.providerFromAddress
	if t.fat != nil {
		if err := t.fat.Execute(buf, params); err != nil {
			t.log.WithFields(logrus.Fields{
				"msg_id":        in.UID,
				logrus.ErrorKey: err,
			}).Error("failed to parse FromAddress")
			return mailprovider2.Message{}, err
		}

		// Get the from address.
		fromAddressStr := buf.String()

		var err error
		fromAddr, err = mail.ParseAddress(fromAddressStr)
		if err != nil {
			t.log.WithFields(logrus.Fields{
				"msg_id":        in.UID,
				logrus.ErrorKey: err,
			}).Error("failed parsing template parsed FromAddress")
			return mailprovider2.Message{}, err
		}
	}

	// Parse the subject.
	if err := t.st.Execute(buf, params); err != nil {
		t.log.WithFields(logrus.Fields{
			"msg_id":        in.UID,
			logrus.ErrorKey: err,
		}).Error("failed to parse template Subject")
		return mailprovider2.Message{}, err
	}

	subject := buf.String()

	// Parse the body.
	if err := t.bt.Execute(buf, params); err != nil {
		t.log.WithFields(logrus.Fields{
			"msg_id":        in.UID,
			logrus.ErrorKey: err,
		}).Error("failed to parse template Body")
		return mailprovider2.Message{}, err
	}

	body := buf.String()

	buf.Reset()

	toAddresses := make([]*mail.Address, 0, len(in.ToAddress))
	for _, toAddr := range in.ToAddress {
		parsed, err := mail.ParseAddress(toAddr)
		if err != nil {
			t.log.WithFields(logrus.Fields{
				"msg_id":  in.UID,
				"to_addr": toAddr,
			}).Error("failed to parse ToAddress")
			return mailprovider2.Message{}, newInvalidAddressError("to_address", toAddr, err)
		}
		toAddresses = append(toAddresses, parsed)
	}

	ct := http.DetectContentType([]byte(body))

	msg := mailprovider2.Message{
		ID:          in.UID,
		From:        fromAddr,
		To:          toAddresses,
		Subject:     subject,
		Body:        body,
		ContentType: ct,
	}
	return msg, nil
}
