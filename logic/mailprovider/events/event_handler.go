package mailproviderevents

import (
	"context"

	"github.com/blockysource/blocky/services/mailing/public/mailingpb"
)

// EventHandler is the interface that handles events from the mailing service.
type EventHandler interface {
	OnEventCurrentMailingProviderUpdated(ctx context.Context, msg *mailingpb.EventCurrentMailingProviderUpdated)
	OnEventCurrentMailingProviderReplaced(ctx context.Context, msg *mailingpb.EventCurrentMailingProviderReplaced)
}
