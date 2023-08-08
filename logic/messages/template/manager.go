package emailtemplate

import (
	"errors"
	"io"
	"net/mail"
	"strings"
	"sync"
	"text/template"

	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/services/mailing/public/mailing"
	mailprovidermanager "github.com/blockysource/mailing/logic/mailprovider/manager"
)

// Manager is the interface that manages email templates.
type Manager struct {
	templates *templateBTree
	pm        *mailprovidermanager.Manager
	log       *logrus.Entry
	cfg       *mailing.TemplateConfig
}

// ReplaceOrInsert replaces or inserts a template definition and creates a new version of it.
func (m *Manager) ReplaceOrInsert(t *TemplateDefinition) (bool, error) {
	var tp TemplateParser
	if err := m.prepareTemplateParser(t, &tp); err != nil {
		return false, err
	}

	_, replaced := m.templates.ReplaceOrInsert(&tp)
	return replaced, nil
}

// Verify verifies the given template definition.
func (m *Manager) Verify(t *TemplateDefinition) error {
	err := m.prepareTemplateParser(t, nil)
	return err
}

// Get returns the template parser with the given UID.
func (m *Manager) Get(uid string) (*TemplateParser, bool) {
	return m.templates.Get(uid)
}

func (m *Manager) prepareTemplateParser(t *TemplateDefinition, tp *TemplateParser) error {
	var providerFromAddress *mail.Address
	cp, ok := m.pm.GetCurrentProvider()
	if ok {
		providerFromAddress = cp.GetDefaultFromAddress()
	}

	params := map[string]string{}
	// Add the predefined parameters from config.
	for _, p := range m.cfg.Parameters {
		params[p.Name] = p.Value
	}

	// Add the default parameter values from the template definition.
	for _, p := range t.Parameters {
		params[p.Name] = p.DefaultValue
	}

	fromAddrTemp, err := template.New("").
		Option("missingkey=error").
		Parse(t.FromAddress)
	if err != nil {
		return newInvalidTemplateError("from_address", err)
	}
	if err = fromAddrTemp.Execute(io.Discard, params); err != nil {
		return newInvalidTemplateError("from_address", m.parseTemplateExecErr(err))
	}

	subTemp, err := template.New("").
		Option("missingkey=error").
		Parse(t.Subject)
	if err != nil {
		return newInvalidTemplateError("subject", err)
	}
	if err = subTemp.Execute(io.Discard, params); err != nil {
		return newInvalidTemplateError("subject", m.parseTemplateExecErr(err))
	}

	bodyTemp, err := template.New("").
		Option("missingkey=error").
		Parse(t.Body)
	if err != nil {
		return newInvalidTemplateError("body", err)
	}
	if err = bodyTemp.Execute(io.Discard, params); err != nil {
		return newInvalidTemplateError("body", m.parseTemplateExecErr(err))
	}

	if tp == nil {
		return nil
	}
	var parameters []Parameter
	for k, v := range params {
		parameters = append(parameters, Parameter{
			Name:         k,
			DefaultValue: v,
		})
	}
	*tp = TemplateParser{
		l:                   sync.RWMutex{},
		base:                *t,
		providerFromAddress: providerFromAddress,
		fat:                 fromAddrTemp,
		st:                  subTemp,
		bt:                  bodyTemp,
		log:                 m.log.WithField("template_uid", t.UID),
		parameters:          parameters,
	}
	return nil
}

// OnEventCurrentMailingProviderReplaced is called when the current mailing provider is replaced.
func (m *Manager) OnEventCurrentMailingProviderReplaced() error {
	cp, ok := m.pm.GetCurrentProvider()
	if !ok {
		return errors.New("no current provider")
	}
	var err error
	m.templates.Ascend(func(tp *TemplateParser) bool {
		if err = tp.ProviderUpdated(cp); err != nil {
			return false
		}
		return true
	})
	return err
}

func (m *Manager) parseTemplateExecErr(err error) error {
	const erPart = "map has no entry for key"
	var ee template.ExecError
	if !errors.As(err, &ee) {
		return err
	}
	errParam := ee.Err.Error()
	idx := strings.Index(errParam, erPart)
	if idx > 0 {
		errParam = errParam[idx+len(erPart)+1:]
		return newMissingParameterError(errParam)
	}
	return err

}
