//go:build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package emailtemplate

import (
	"github.com/google/wire"
	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/pkg/go/providers"
	"github.com/blockysource/blocky/services/mailing/internal/deps"
	"github.com/blockysource/blocky/services/mailing/public/mailing"
	mailprovidermanager "github.com/blockysource/mailing/logic/mailprovider/manager"
)

// NewManager creates a new Manager.
func NewManager(d *mailing.Dependencies, pm *mailprovidermanager.Manager) (*Manager, error) {
	wire.Build(
		newTemplateBTree,
		deps.GetLogrusLogger,
		wire.Value(deps.ModuleName),
		wire.Value(logrus.Fields{
			"part": "email-template-manager",
		}),
		providers.FieldsLogrusEntry,

		deps.GetConfig,
		deps.GetTemplateConfig,
		wire.Struct(new(Manager), "*"),
	)
	return nil, nil
}
