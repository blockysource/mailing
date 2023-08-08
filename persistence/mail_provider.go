package persistence

import (
	"context"
	"errors"
	"net/mail"

	"github.com/blockysource/blocky/services/mailing/public/mailingpb"
	"github.com/blockysource/mailing/logic/mailprovider"
)

var (
	ErrAlreadyExists       = errors.New("already exists")
	ErrNotFound            = errors.New("not found")
	ErrProviderNotVerified = errors.New("provider not verified")
	ErrInvalidConfigType   = errors.New("invalid config type")
)

// MailingProviderStorage is an interface that represents a mailing provider storage.
type MailingProviderStorage interface {
	CreateProvider(ctx context.Context, in *CreateMailingProviderArgs) (mailprovider.MailingProviderDefinition, error)
	UpdateProvider(ctx context.Context, in *UpdateMailingProviderArgs) (UpdateMailingProviderResult, error)
	SetCurrentProvider(ctx context.Context, in *SetCurrentMailingProviderArgs) error
	MarkProviderVerified(ctx context.Context, in *MarkProviderVerifiedArgs) error
	GetCurrentProvider(ctx context.Context) (mailprovider.MailingProviderDefinition, error)
	ListProviders(ctx context.Context) ([]mailprovider.MailingProviderDefinition, error)
}

// CreateMailingProviderArgs creates a new mailing provider.
type CreateMailingProviderArgs struct {
	// UID is the unique identifier of the mailing provider.
	UID string
	// Name is the name of the mailing provider.
	Name string
	// FromAddress is the default from address of the mailing provider.
	FromAddress *mail.Address
	// Type is the type of the mailing provider.
	Type mailingpb.MailingProviderType
	// Config is the configuration of the mailing provider.
	Config mailingpb.MailingProviderConfig
}
type (
	// UpdateMailingProviderArgs updates a mailing provider.
	UpdateMailingProviderArgs struct {
		// UID is the unique identifier of the mailing provider.
		UID string
		// Name is the name of the mailing provider.
		Name string
		// FromAddress is the default from address of the mailing provider.
		FromAddress *mail.Address
		// Config is the configuration of the mailing provider.
		Config *mailingpb.MailingProviderConfig
	}
	// UpdateMailingProviderResult is the result of the UpdateMailingProvider method.
	UpdateMailingProviderResult struct {
		WasInUse        bool
		MailingProvider mailprovider.MailingProviderDefinition
	}
)

// SetCurrentMailingProviderArgs sets the current mailing provider.
type SetCurrentMailingProviderArgs struct {
	// UID is the unique identifier of the mailing provider.
	UID string
}
