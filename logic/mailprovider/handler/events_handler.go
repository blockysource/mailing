package mailproviderhandler

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/pkg/go/providers"
	"github.com/blockysource/blocky/services/mailing/internal/persistence"
	"github.com/blockysource/blocky/services/mailing/public/mailingpb"
	mailprovidermanager "github.com/blockysource/mailing/logic/mailprovider/manager"
)

// EventsHandler is the structure that handles events from the mailing service.
type EventsHandler struct {
	l   sync.RWMutex `wire:"-"`
	m   *mailprovidermanager.Manager
	p   persistence.MailingProviderStorage
	n   providers.ServiceNonce
	log *logrus.Entry
}

// OnEventCurrentMailingProviderUpdated handles the event of the current mailing provider being updated.
func (e *EventsHandler) OnEventCurrentMailingProviderUpdated(ctx context.Context, msg *mailingpb.EventCurrentMailingProviderUpdated) {
	if e.n == msg.Nonce {
		e.log.Trace("skipping event EventCurrentMailingProviderUpdated, nonce is the same")
		return
	}
	e.l.Lock()
	defer e.l.Unlock()

	cp, ok := e.m.GetCurrentProvider()
	if !ok {
		e.log.Trace("skipping event EventCurrentMailingProviderUpdated, no current provider")
		return
	}

	// Current Provider ID should match the one in the event.
	if cp.GetID() != msg.UID {
		e.log.WithFields(logrus.Fields{
			"current_provider_id": cp.GetID(),
			"event_provider_id":   msg.UID,
			"nonce":               msg.Nonce,
		}).Warn("skipping event EventCurrentMailingProviderUpdated, current provider ID does not match the one in the event")
	}

	e.m.UnsetCurrentProvider()
	e.log.WithFields(logrus.Fields{
		"current_provider_id": cp.GetID(),
		"nonce":               msg.Nonce,
	}).Debug("updated current mail provider unset successfully")

	// Close the current provider.
	cp.Close()
}

// OnEventCurrentMailingProviderReplaced handles the event of the current mailing provider being replaced.
func (e *EventsHandler) OnEventCurrentMailingProviderReplaced(ctx context.Context, msg *mailingpb.EventCurrentMailingProviderReplaced) {
	if e.n == msg.Nonce {
		e.log.Trace("skipping event EventCurrentMailingProviderReplaced, nonce is the same")
		return
	}
	e.l.Lock()
	defer e.l.Unlock()

	oldProvider, ok := e.m.GetCurrentProvider()
	if ok {
		defer oldProvider.Close()
	}

	mp, err := e.p.GetCurrentProvider(ctx)
	if err != nil {
		e.log.WithError(err).Error("failed to get current mail provider")
		return
	}

	p, err := e.m.LoadProvider(ctx, mp)
	if err != nil {
		e.log.WithError(err).Error("failed to load current mail provider")
		return
	}

	// Replace the current provider.
	if err = e.m.ReplaceCurrentProvider(p); err != nil {
		e.log.WithError(err).Error("failed to replace current mail provider")
		return
	}
	e.log.WithFields(logrus.Fields{
		"current_provider_id": p.GetID(),
		"nonce":               msg.Nonce,
	}).Debug("updated current mail provider replaced successfully")
}
