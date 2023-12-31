// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package natsmailproviderevents

import (
	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/pkg/go/providers"
	"github.com/blockysource/blocky/services/mailing/internal/deps"
	"github.com/blockysource/blocky/services/mailing/public/mailing"
	mailproviderevents "github.com/blockysource/mailing/logic/mailprovider/events"
)

// Injectors from wire.go:

// NewPublisher creates a new event Publisher.
func NewPublisher(dependencies *mailing.Dependencies) (*Publisher, error) {
	conn, err := deps.GetNatsConn(dependencies)
	if err != nil {
		return nil, err
	}
	moduleName := _wireModuleNameValue
	logger, err := deps.GetLogrusLogger(dependencies)
	if err != nil {
		return nil, err
	}
	fields := _wireFieldsValue
	entry, err := providers.FieldsLogrusEntry(moduleName, logger, fields)
	if err != nil {
		return nil, err
	}
	config, err := deps.GetConfig(dependencies)
	if err != nil {
		return nil, err
	}
	eventsConfig, err := deps.GetEventsConfig(config)
	if err != nil {
		return nil, err
	}
	publisher := &Publisher{
		nc:  conn,
		log: entry,
		cfg: eventsConfig,
	}
	return publisher, nil
}

var (
	_wireModuleNameValue = deps.ModuleName
	_wireFieldsValue     = logrus.Fields{
		"part": "natsmailproviderevents",
		"type": "publisher",
	}
)

// NewListener creates a new event Listener.
func NewListener(dependencies *mailing.Dependencies, eventHandler mailproviderevents.EventHandler) (*Listener, error) {
	conn, err := deps.GetNatsConn(dependencies)
	if err != nil {
		return nil, err
	}
	config, err := deps.GetConfig(dependencies)
	if err != nil {
		return nil, err
	}
	eventsConfig, err := deps.GetEventsConfig(config)
	if err != nil {
		return nil, err
	}
	moduleName := _wireProvidersModuleNameValue
	logger, err := deps.GetLogrusLogger(dependencies)
	if err != nil {
		return nil, err
	}
	fields := _wireLogrusFieldsValue
	entry, err := providers.FieldsLogrusEntry(moduleName, logger, fields)
	if err != nil {
		return nil, err
	}
	listener := &Listener{
		nc:  conn,
		eh:  eventHandler,
		cfg: eventsConfig,
		log: entry,
	}
	return listener, nil
}

var (
	_wireProvidersModuleNameValue = deps.ModuleName
	_wireLogrusFieldsValue        = logrus.Fields{
		"part": "natsmailproviderevents",
		"type": "listener",
	}
)
