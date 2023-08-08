package natsmailproviderevents

import (
	"context"
	"sync/atomic"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/services/mailing/public/mailing"
	"github.com/blockysource/blocky/services/mailing/public/mailingpb"
	mailproviderevents "github.com/blockysource/mailing/logic/mailprovider/events"
)

// Listener is a mail provider event listener on NATS.
type Listener struct {
	nc  *nats.Conn
	eh  mailproviderevents.EventHandler
	cfg *mailing.EventsConfig
	log *logrus.Entry

	closeFn   context.CancelFunc `wire:"-"`
	isStarted atomic.Bool        `wire:"-"`
}

// Start listening for the NATS mail provider events.
func (l *Listener) Start(ctx context.Context) error {
	if !l.isStarted.CompareAndSwap(false, true) {
		l.log.Warn("listener already started")
		return nil
	}
	ctx, l.closeFn = context.WithCancel(ctx)
	if err := l.listenAndHandle(ctx); err != nil {
		return err
	}
	return nil
}

// Stop listening for the NATS mail provider events.
func (l *Listener) Stop() error {
	if !l.isStarted.CompareAndSwap(true, false) {
		l.log.Warn("listener already stopped")
		return nil
	}
	l.closeFn()
	return nil
}

func (l *Listener) listenAndHandle(ctx context.Context) error {
	l.log.Debug("starting to listen and handle events")
	// Listen and handle events.
	// 1. Current Mail Provider Updated.
	if err := l.listenOnCurrentMailProviderUpdated(ctx); err != nil {
		return err
	}
	// 2. Current Mail Provider Replaced.
	if err := l.listenOnCurrentMailProviderReplaced(ctx); err != nil {
		return err
	}
	return nil
}

func (l *Listener) listenOnCurrentMailProviderUpdated(ctx context.Context) error {
	sub, err := l.nc.SubscribeSync(mailing.EventCurrentMailingProviderUpdatedTopic(l.cfg.Prefix))
	if err != nil {
		return err
	}

	go func(ctx context.Context, sub *nats.Subscription) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var msg *nats.Msg
				msg, err = sub.NextMsgWithContext(ctx)
				if err != nil {
					l.log.
						WithField("topic", mailing.EventCurrentMailingProviderUpdatedTopic(l.cfg.Prefix)).
						WithError(err).Error("failed to get next message, stopping listener")
					return
				}
				var event mailingpb.EventCurrentMailingProviderUpdated
				if err = event.Unmarshal(msg.Data); err != nil {
					l.log.WithFields(logrus.Fields{
						"topic":         mailing.EventCurrentMailingProviderUpdatedTopic(l.cfg.Prefix),
						logrus.ErrorKey: err,
					}).Error("failed to unmarshal event, skipping")
					continue
				}
				l.eh.OnEventCurrentMailingProviderUpdated(ctx, &event)
			}
		}
	}(ctx, sub)
	return nil
}

func (l *Listener) listenOnCurrentMailProviderReplaced(ctx context.Context) error {
	sub, err := l.nc.SubscribeSync(mailing.EventCurrentMailingProviderReplacedTopic(l.cfg.Prefix))
	if err != nil {
		return err
	}

	go func(ctx context.Context, sub *nats.Subscription) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var msg *nats.Msg
				msg, err = sub.NextMsgWithContext(ctx)
				if err != nil {
					l.log.
						WithField("topic", mailing.EventCurrentMailingProviderReplacedTopic(l.cfg.Prefix)).
						WithError(err).Error("failed to get next message, stopping listener")
					return
				}
				var event mailingpb.EventCurrentMailingProviderReplaced
				if err = event.Unmarshal(msg.Data); err != nil {
					l.log.WithFields(logrus.Fields{
						"topic":         mailing.EventCurrentMailingProviderReplacedTopic(l.cfg.Prefix),
						logrus.ErrorKey: err,
					}).Error("failed to unmarshal event, skipping")
					continue
				}
				l.eh.OnEventCurrentMailingProviderReplaced(ctx, &event)
			}
		}
	}(ctx, sub)
	return nil
}
