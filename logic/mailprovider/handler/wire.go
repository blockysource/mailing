//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package mailproviderhandler

import (
	"github.com/google/wire"
	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/pkg/go/providers"
	"github.com/blockysource/blocky/services/mailing/internal/deps"
	"github.com/blockysource/blocky/services/mailing/internal/persistence"
	"github.com/blockysource/blocky/services/mailing/public/mailing"
	mailproviderevents "github.com/blockysource/mailing/logic/mailprovider/events"
	mailprovidermanager "github.com/blockysource/mailing/logic/mailprovider/manager"
)

// NewHandler creates a new Handler.
func NewHandler(*mailing.Dependencies, mailproviderevents.Publisher, persistence.MailingProviderStorage, providers.ServiceNonce) (*Handler, error) {
	wire.Build(
		// Logger.
		deps.GetLogrusLogger,
		wire.Value(deps.ModuleName),
		wire.Value(logrus.Fields{
			"part": "mailproviderhandler",
			"type": "handler",
		}),
		providers.FieldsLogrusEntry,
		wire.Struct(new(Handler), "*"),
	)
	return nil, nil
}

// NewEventsHandler creates a new EventsHandler.
func NewEventHandler(*mailing.Dependencies, persistence.MailingProviderStorage, *mailprovidermanager.Manager, providers.ServiceNonce) (*EventsHandler, error) {
	wire.Build(
		// Logger.
		deps.GetLogrusLogger,
		wire.Value(deps.ModuleName),
		wire.Value(logrus.Fields{
			"part": "mailproviderhandler",
			"type": "eventhandler",
		}),
		providers.FieldsLogrusEntry,
		wire.Struct(new(EventsHandler), "*"),
	)
	return nil, nil
}
