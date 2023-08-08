//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package mailprovidermanager

import (
	"github.com/google/wire"
	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/pkg/go/providers"
	"github.com/blockysource/blocky/services/mailing/internal/deps"
	"github.com/blockysource/blocky/services/mailing/public/mailing"
	smtpprovider "github.com/blockysource/mailing/logic/mailprovider/smtp"
)

// New creates a new Manager.
func New(d *mailing.Dependencies) (*Manager, error) {
	wire.Build(
		deps.GetConfig,
		deps.GetLogrusLogger,
		wire.Value(deps.ModuleName),
		wire.Value(logrus.Fields{
			"part": "provider-manager",
		}),
		providers.FieldsLogrusEntry,
		smtpprovider.NewSMTPProvidersConfig,
		wire.Struct(new(Manager), "*"),
	)
	return nil, nil
}
