// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package mailprovidermanager

import (
	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/pkg/go/providers"
	"github.com/blockysource/blocky/services/mailing/internal/deps"
	"github.com/blockysource/blocky/services/mailing/public/mailing"
	smtpmailprovider "github.com/blockysource/mailing/logic/mailprovider/smtp"
)

// Injectors from wire.go:

// New creates a new Manager.
func New(d *mailing.Dependencies) (*Manager, error) {
	moduleName := _wireModuleNameValue
	logger, err := deps.GetLogrusLogger(d)
	if err != nil {
		return nil, err
	}
	fields := _wireFieldsValue
	entry, err := providers.FieldsLogrusEntry(moduleName, logger, fields)
	if err != nil {
		return nil, err
	}
	config, err := deps.GetConfig(d)
	if err != nil {
		return nil, err
	}
	smtpProvidersConfig, err := smtpmailprovider.NewSMTPProvidersConfig(config, entry)
	if err != nil {
		return nil, err
	}
	manager := &Manager{
		log: entry,
		sc:  smtpProvidersConfig,
	}
	return manager, nil
}

var (
	_wireModuleNameValue = deps.ModuleName
	_wireFieldsValue     = logrus.Fields{
		"part": "provider-manager",
	}
)