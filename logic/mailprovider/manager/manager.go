package mailprovidermanager

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"sync"

	"github.com/google/uuid"
	"github.com/pallinder/go-randomdata"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mailingadminv1 "github.com/blockysource/go-genproto/blockyapis/mailing/admin/v1alpha"
	"github.com/blockysource/mailing/logic/mailprovider"
	smtpmailprovider "github.com/blockysource/mailing/logic/mailprovider/smtp"
)

// Manager is responsible for managing currently used provider.
type Manager struct {
	l sync.RWMutex `wire:"-"`
	// current is a currently used mailprovider.Provider
	current mailprovider.Provider `wire:"-"`

	log *slog.Logger
	sc  *smtpmailprovider.SMTPProvidersConfig
}

// NewProviderBase creates a new provider from the given configuration.
func (m *Manager) NewProviderBase(in *mailingadminv1.CreateMailingProviderRequest) (mailprovider.Base, error) {
	switch in.Type {
	case mailingadminv1.MailingProviderType_SMTP:
		// Verify that the input config is valid - the validate function checks if the input is nil.
		if err := in.Config.GetSmtpConfig().Validate(); err != nil {
			return mailprovider.Base{}, err
		}
	default:
		return mailprovider.Base{}, status.Error(codes.Unimplemented, "not implemented yet")
	}

	b := mailprovider.Base{
		ID:   in.UID,
		Type: in.Type,
		Name: in.Name,
	}

	// Check if the unique identifier was provided and generate a random one if it is empty.
	if b.ID == "" {
		b.ID = uuid.New().String()
	}

	// Check if the name is empty and generate a random one if it is.
	if b.Name == "" {
		b.Name = fmt.Sprintf("%s %s", in.Type.String(), randomdata.SillyName())
	}

	addr, err := mail.ParseAddress(in.FromAddress)
	if err != nil {
		return mailprovider.Base{}, status.Error(codes.InvalidArgument, "invalid from address")
	}
	b.FromAddress = addr

	return b, nil
}

// LoadProvider loads a provider from the given configuration.
func (m *Manager) LoadProvider(ctx context.Context, def mailprovider.MailingProviderDefinition) (mailprovider.Provider, error) {
	switch def.Type {
	case mailingadminv1.SMTP:
		return smtpmailprovider.New(m.sc, def, m.log)
	default:
		return nil, status.Error(codes.Unimplemented, "not implemented yet")
	}
}

// ReplaceCurrentProvider replaces the current provider with the given one.
func (m *Manager) ReplaceCurrentProvider(p mailprovider.Provider) error {
	if !p.IsVerified() {
		return errors.New("cannot replace current provider - it is not verified")
	}

	m.l.Lock()
	m.current = p
	m.l.Unlock()
	return nil
}

// GetCurrentProvider returns current mailprovider.Provider
func (m *Manager) GetCurrentProvider() (mailprovider.Provider, bool) {
	m.l.RLock()
	defer m.l.RUnlock()
	return m.current, m.current != nil
}

// UnsetCurrentProvider unsets the current provider.
func (m *Manager) UnsetCurrentProvider() {
	m.l.Lock()
	if m.current != nil {
		m.current.Close()
	}
	m.current = nil
	m.l.Unlock()
}
