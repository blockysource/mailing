package mailproviderhandler

import (
	"context"
	"errors"
	"net/mail"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/blockysource/blocky/pkg/go/providers"
	"github.com/blockysource/blocky/services/mailing/internal/persistence"
	"github.com/blockysource/blocky/services/mailing/public/mailingpb"
	mailproviderevents "github.com/blockysource/mailing/logic/mailprovider/events"
	mailprovidermanager "github.com/blockysource/mailing/logic/mailprovider/manager"
)

var _ mailingpb.MailingProviderServiceServer = (*Handler)(nil)

// Handler is a handler that handles emails.
type Handler struct {
	p   persistence.MailingProviderStorage
	m   *mailprovidermanager.Manager
	log *logrus.Entry
	ep  mailproviderevents.Publisher
	n   providers.ServiceNonce
}

// UpdateMailingProvider updates a mailing provider.
func (h *Handler) UpdateMailingProvider(ctx context.Context, in *mailingpb.UpdateMailingProviderRequest) (*mailingpb.UpdateMailingProviderResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var fromAddress *mail.Address
	if in.FromAddress != "" {
		var err error
		fromAddress, err = mail.ParseAddress(in.FromAddress)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid from address %s", err.Error())
		}
	}

	args := persistence.UpdateMailingProviderArgs{
		UID:         in.UID,
		Name:        in.Name,
		FromAddress: fromAddress,
		Config:      in.Config,
	}

	// Update the provider in the persistence layer.
	out, err := h.p.UpdateProvider(ctx, &args)
	if err != nil {
		if errors.Is(err, persistence.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "mailing provider not found")
		}
		return nil, err
	}

	// If the provider was in use, publish a message to the event publisher.
	if out.WasInUse {
		h.log.WithContext(ctx).
			WithFields(logrus.Fields{
				"uid":  in.UID,
				"name": in.Name,
			}).Info("current mailing provider updated, waiting for verification")

		// If the provider was the current provider, unset the current provider.
		// The provider needs to wait for the verification now.
		h.m.UnsetCurrentProvider()

		msg := mailingpb.EventCurrentMailingProviderUpdated{
			UID:   in.UID,
			Nonce: h.n,
		}
		if err = h.ep.PublishCurrentMailingProviderUpdated(ctx, &msg); err != nil {
			h.log.WithContext(ctx).WithError(err).Error("failed to publish event")
		}
		h.log.WithContext(ctx).
			WithField("uid", in.UID).
			Trace("current mailing provider updated event published")
	}

	h.log.WithContext(ctx).
		WithField("uid", in.UID).
		Debug("mailing provider updated")

	return &mailingpb.UpdateMailingProviderResponse{MailingProvider: out.MailingProvider.ToProto()}, nil
}

// CreateMailingProvider creates a new mailing provider.
func (h *Handler) CreateMailingProvider(ctx context.Context, in *mailingpb.CreateMailingProviderRequest) (*mailingpb.CreateMailingProviderResponse, error) {
	// validate the request.
	if err := in.Validate(); err != nil {
		if se, ok := status.FromError(err); ok {
			return nil, se.Err()
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Parse and validate the input into provider.Base.
	base, err := h.m.NewProviderBase(in)
	if err != nil {
		h.log.WithContext(ctx).WithError(err).Error("failed to create mailing provider")
		return nil, err
	}

	// Define arguments for the persistence layer.
	args := persistence.CreateMailingProviderArgs{
		UID:         base.ID,
		Name:        base.Name,
		FromAddress: base.FromAddress,
		Type:        base.Type,
		Config:      in.Config,
	}

	// Store the provider in the persistence layer.
	out, err := h.p.CreateProvider(ctx, &args)
	if err != nil {
		if errors.Is(err, persistence.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "mailing provider already exists")
		}
		return nil, err
	}

	return &mailingpb.CreateMailingProviderResponse{
		MailingProvider: mailingpb.MailingProvider{
			UID:         out.UID,
			Name:        out.Name,
			Type:        out.Type,
			InUse:       false,
			Config:      in.Config,
			CreatedAt:   out.CreatedAt,
			VerifiedAt:  nil,
			FromAddress: out.FromAddress,
		},
	}, nil
}

// SetCurrentMailingProvider sets the current mailing provider.
func (h *Handler) SetCurrentMailingProvider(ctx context.Context, in *mailingpb.SetCurrentMailingProviderRequest) (*mailingpb.SetCurrentMailingProviderResponse, error) {
	// validate the request.
	if err := in.Validate(); err != nil {
		if se, ok := status.FromError(err); ok {
			return nil, se.Err()
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Set current provider in the persistence layer.
	args := persistence.SetCurrentMailingProviderArgs{UID: in.UID}
	if err := h.p.SetCurrentProvider(ctx, &args); err != nil {
		if errors.Is(err, persistence.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "mailing provider not found")
		}
		return nil, err
	}

	// Publish the event.
	msg := mailingpb.EventCurrentMailingProviderReplaced{NewUID: in.UID, Nonce: h.n}
	if err := h.ep.PublishCurrentMailingProviderReplaced(ctx, &msg); err != nil {
		return nil, err

	}

	return &mailingpb.SetCurrentMailingProviderResponse{}, nil
}

// ListMailingProviders lists all mailing providers.
func (h *Handler) ListMailingProviders(ctx context.Context, in *mailingpb.ListMailingProvidersRequest) (*mailingpb.ListMailingProvidersResponse, error) {
	ls, err := h.p.ListProviders(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list mailing providers")
	}

	var providers []mailingpb.MailingProvider
	for _, p := range ls {
		providers = append(providers, p.ToProto())
	}

	return &mailingpb.ListMailingProvidersResponse{
		MailingProviders: providers,
	}, nil
}

// GetCurrentMailingProvider gets the current mailing provider.
func (h *Handler) GetCurrentMailingProvider(ctx context.Context, in *mailingpb.GetCurrentMailingProviderRequest) (*mailingpb.GetCurrentMailingProviderResponse, error) {
	p, ok := h.m.GetCurrentProvider()
	if !ok {
		return nil, status.Error(codes.NotFound, "no current mailing provider")
	}

	pd := p.GetDefinition()

	var verifiedAt *time.Time
	if !pd.VerifiedAt.IsZero() {
		verifiedAt = &pd.VerifiedAt
	}
	return &mailingpb.GetCurrentMailingProviderResponse{
		MailingProvider: mailingpb.MailingProvider{
			UID:         pd.UID,
			Name:        pd.Name,
			Type:        pd.Type,
			InUse:       true,
			Config:      *pd.Config.Clone(),
			CreatedAt:   pd.CreatedAt,
			VerifiedAt:  verifiedAt,
			FromAddress: pd.FromAddress,
		},
	}, nil

}
