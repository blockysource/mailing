package natsmailproviderevents

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/services/mailing/public/mailing"
	"github.com/blockysource/blocky/services/mailing/public/mailingpb"
	mailproviderevents "github.com/blockysource/mailing/logic/mailprovider/events"
)

var _ mailproviderevents.Publisher = (*Publisher)(nil)

type Publisher struct {
	nc  *nats.Conn
	log *logrus.Entry
	cfg *mailing.EventsConfig
}

// PublishCurrentMailingProviderUpdated publishes a provider updated event.
func (p *Publisher) PublishCurrentMailingProviderUpdated(ctx context.Context, in *mailingpb.EventCurrentMailingProviderUpdated) error {
	if err := in.Validate(); err != nil {
		p.log.WithError(err).Error("failed to validate current mailing provider updated event")
		return err
	}

	// Marshal the message.
	data, err := in.Marshal()
	if err != nil {
		p.log.WithError(err).Error("failed to marshal current mailing provider updated event")
		return err
	}

	// Publish the message.
	if err = p.nc.Publish(mailing.EventCurrentMailingProviderUpdatedTopic(p.cfg.Prefix), data); err != nil {
		p.log.WithError(err).Error("failed to publish current mailing provider updated event")
		return err
	}

	p.log.Trace("published current mailing provider updated event")
	return nil
}

func (p *Publisher) PublishCurrentMailingProviderReplaced(ctx context.Context, in *mailingpb.EventCurrentMailingProviderReplaced) error {
	if err := in.Validate(); err != nil {
		p.log.WithError(err).Error("failed to validate current mailing provider changed event")
		return err
	}
	// Marshal the message.
	data, err := in.Marshal()
	if err != nil {
		p.log.WithError(err).Error("failed to marshal current mailing provider changed event")
		return err
	}

	// Publish the message.
	if err = p.nc.Publish(mailing.EventCurrentMailingProviderReplacedTopic(p.cfg.Prefix), data); err != nil {
		p.log.WithError(err).Error("failed to publish current mailing provider changed event")
		return err
	}

	p.log.Trace("published current mailing provider changed event")
	return nil
}
