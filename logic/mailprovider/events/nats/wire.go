//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package natsmailproviderevents

import (
	"github.com/google/wire"
	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/pkg/go/providers"
	"github.com/blockysource/blocky/services/mailing/internal/deps"
	"github.com/blockysource/blocky/services/mailing/public/mailing"
	mailproviderevents "github.com/blockysource/mailing/logic/mailprovider/events"
)

// NewPublisher creates a new event Publisher.
func NewPublisher(*mailing.Dependencies) (*Publisher, error) {
	wire.Build(
		deps.GetConfig,
		deps.GetEventsConfig,
		deps.GetLogrusLogger,
		wire.Value(deps.ModuleName),
		wire.Value(logrus.Fields{
			"part": "natsmailproviderevents",
			"type": "publisher",
		}),
		providers.FieldsLogrusEntry,
		deps.GetNatsConn,
		wire.Struct(new(Publisher), "*"),
	)
	return nil, nil
}

// NewListener creates a new event Listener.
func NewListener(*mailing.Dependencies, mailproviderevents.EventHandler) (*Listener, error) {
	wire.Build(
		deps.GetConfig,
		deps.GetEventsConfig,
		deps.GetLogrusLogger,
		wire.Value(deps.ModuleName),
		wire.Value(logrus.Fields{
			"part": "natsmailproviderevents",
			"type": "listener",
		}),
		providers.FieldsLogrusEntry,
		deps.GetNatsConn,
		wire.Struct(new(Listener), "*"),
	)
	return nil, nil
}
