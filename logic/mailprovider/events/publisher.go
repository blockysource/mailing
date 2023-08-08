package mailproviderevents

import (
	"context"

	"github.com/blockysource/blocky/services/mailing/public/mailingpb"
)

type Publisher interface {
	// PublishCurrentMailingProviderReplaced publishes a provider created event.
	PublishCurrentMailingProviderReplaced(ctx context.Context, in *mailingpb.EventCurrentMailingProviderReplaced) error
	// PublishCurrentMailingProviderUpdated publishes a provider updated event.
	PublishCurrentMailingProviderUpdated(ctx context.Context, in *mailingpb.EventCurrentMailingProviderUpdated) error
}
