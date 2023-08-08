package smtpmailprovider

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/open-source/libs/blocky-cloud/email/message"

	"github.com/blockysource/blocky/services/mailing/public/mailingpb"
	mailprovider2 "github.com/blockysource/mailing/logic/mailprovider"
)

var _bufioWriterPool = sync.Pool{}

func newBufioWriter(w io.Writer) *bufio.Writer {
	if v := _bufioWriterPool.Get(); v != nil {
		bw := v.(*bufio.Writer)
		bw.Reset(w)
		return bw
	}
	return bufio.NewWriter(w)
}

func putBufioWriter(bw *bufio.Writer) {
	bw.Reset(nil)
	_bufioWriterPool.Put(bw)
}

// Compile-time check to verify implements interface.
var _ mailprovider2.Provider = (*SMTPProvider)(nil)

// SMTPProvider is a provider that sends emails via SMTP.
type SMTPProvider struct {
	l sync.RWMutex

	pc          *SMTPProvidersConfig
	p           mailprovider2.MailingProviderDefinition
	fromAddress *mail.Address
	cfg         mailingpb.SMTPConfig
	log         *logrus.Entry
	isVerified  bool
}

// New creates a new SMTP provider.
func New(mc *SMTPProvidersConfig, p mailprovider2.MailingProviderDefinition, log *logrus.Entry) (*SMTPProvider, error) {
	// Get the configuration.
	cfg := p.Config.GetSmtpConfig()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &SMTPProvider{
		pc:  mc,
		p:   p,
		cfg: *cfg,
		log: log.WithFields(logrus.Fields{
			"provider_id": p.UID,
			"provider":    mailingpb.SMTP,
		}),
	}, nil
}

// Close closes the provider.
func (s *SMTPProvider) Close() {}

// GetID returns the ID of the provider.
func (s *SMTPProvider) GetID() string {
	return s.p.UID
}

// GetDefinition returns the provider definition.
func (s *SMTPProvider) GetDefinition() mailprovider2.MailingProviderDefinition {
	return s.p
}

// Type returns the type of the provider.
func (s *SMTPProvider) Type() mailingpb.MailingProviderType {
	return mailingpb.SMTP
}

// IsVerified returns whether the provider is verified.
func (s *SMTPProvider) IsVerified() bool {
	s.l.RLock()
	defer s.l.RUnlock()

	return s.isVerified
}

// UpdateConfig updates the config of the provider.
func (s *SMTPProvider) UpdateConfig(config *mailingpb.MailingProviderConfig) error {
	cfg := config.GetSmtpConfig()
	if cfg == nil {
		return fmt.Errorf("invalid config type: %T", config)
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	s.l.Lock()
	s.cfg = *cfg
	s.l.Unlock()

	return nil
}

func (s *SMTPProvider) GetConfig() mailingpb.MailingProviderConfig {
	s.l.RLock()
	defer s.l.RUnlock()

	cfg := s.cfg
	return mailingpb.MailingProviderConfig{
		Config: &mailingpb.MailingProviderConfig_SmtpConfig{
			SmtpConfig: &cfg,
		},
	}
}

// GetDefaultFromAddress returns the default from address.
func (s *SMTPProvider) GetDefaultFromAddress() *mail.Address {
	s.l.RLock()
	defer s.l.RUnlock()

	return s.fromAddress
}

// Send lets the provider send the input message.
func (s *SMTPProvider) Send(ctx context.Context, msg *message.Message) error {
	// Send the message via SMTP.
	if err := s.sendMessage(ctx, msg); err != nil {
		return err
	}

	s.log.WithFields(logrus.Fields{
		"msg_id":      msg.ID,
		"provider":    mailingpb.SMTP,
		"provider_id": s.p.UID,
	}).Debug("email message sent successfully")
	return nil
}

// Verify verifies the provider configuration.
func (s *SMTPProvider) Verify(ctx context.Context) error {
	s.l.Lock()
	defer s.l.Unlock()

	c, err := s.smtpClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	// Do the hello.
	if err = c.Hello(s.pc.Domain); err != nil {
		s.log.WithFields(logrus.Fields{
			"provider":    mailingpb.SMTP,
			"provider_id": s.p.UID,
		}).Debug("failed to say hello to smtp client")
		return fmt.Errorf("failed to say hello to smtp client: %w", err)
	}

	// Check the TLS.
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err = c.StartTLS(&tls.Config{
			ServerName: s.cfg.Host.UnsafeString(),
		}); err != nil {
			s.log.WithFields(logrus.Fields{
				"provider":      mailingpb.SMTP,
				"provider_id":   s.p.UID,
				logrus.ErrorKey: err,
			}).Debug("failed to start tls on smtp client")
			return fmt.Errorf("failed to start tls: %w", err)
		}
	}

	// Check the authentication of the smtp server.
	if err = c.Auth(s.auth()); err != nil {
		s.log.WithFields(logrus.Fields{
			"provider":    mailingpb.SMTP,
			"provider_id": s.p.UID,
		}).Debug("failed to authenticate smtp client")
		return fmt.Errorf("failed to authenticate smtp client: %w", err)
	}

	if err = c.Quit(); err != nil {
		return fmt.Errorf("failed to quit smtp client: %w", err)
	}

	s.isVerified = true
	return nil
}

func (s *SMTPProvider) smtpClient(ctx context.Context) (*smtp.Client, error) {
	d := net.Dialer{Timeout: 10 * time.Second}

	addr := s.addr()

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to split host and port: %w", err)
	}

	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial smtp server: %w", err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to dial smtp server: %w", err)
	}
	return c, nil
}

func (s *SMTPProvider) sendMessage(ctx context.Context, msg *message.Message) error {
	c, err := s.smtpClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	// Call Hello to the smtp server.
	if err = c.Hello(s.pc.Domain); err != nil {
		s.log.WithError(err).Debug("failed to say hello")
		return errors.New("failed to say hello")
	}

	if ok, _ := c.Extension("STARTTLS"); ok {
		config := tls.Config{ServerName: string(s.cfg.Host)}
		if err = c.StartTLS(&config); err != nil {
			s.log.
				WithFields(logrus.Fields{
					"provider_id": s.p.UID,
				}).WithError(err).Debug("failed to start tls for smtp client")
			return err
		}
	}

	if ok, _ := c.Extension("AUTH"); !ok {
		s.log.WithFields(logrus.Fields{
			"provider_id": s.p.UID,
			"host":        s.cfg.Host,
		}).Debug("smtp server does not support authentication")
		return mailprovider2.ErrAuth(errors.New("smtp server does not support authentication"))
	}

	// Authenticate the smtp client.
	if err = c.Auth(s.auth()); err != nil {
		s.log.WithError(err).Debug("failed to authenticate smtp client")
		return mailprovider2.ErrAuth(err)
	}

	// Set the sender.
	if err = c.Mail(msg.From.String()); err != nil {
		s.log.WithFields(logrus.Fields{
			"provider_id":   s.p.UID,
			"msg_id":        msg.ID,
			"from":          msg.From.String(),
			logrus.ErrorKey: err,
		}).Debug("failed to set sender")

		return s.handleErr(err)
	}

	for _, to := range msg.To {
		// Set the recipients.
		if err = c.Rcpt(to.String()); err != nil {
			s.log.
				WithFields(logrus.Fields{
					"provider_id":   s.p.UID,
					"msg_id":        msg.ID,
					"to":            to.String(),
					logrus.ErrorKey: err,
				}).Debug("failed to set recipient")
			return s.handleErr(err)
		}
	}

	w, err := c.Data()
	if err != nil {
		s.log.
			WithFields(logrus.Fields{
				"provider_id":   s.p.UID,
				"msg_id":        msg.ID,
				logrus.ErrorKey: err,
			}).Debug("failed to get data writer")
		return s.handleErr(err)
	}

	bw := newBufioWriter(w)
	defer putBufioWriter(bw)

	if err = s.writeMessage(msg, bw); err != nil {
		s.log.WithFields(logrus.Fields{
			"provider_id":   s.p.UID,
			"msg_id":        msg.ID,
			logrus.ErrorKey: err,
		}).Debug("failed to write message")
		return errors.New("failed to write message")
	}

	if err = bw.Flush(); err != nil {
		s.log.WithFields(logrus.Fields{
			"provider_id":   s.p.UID,
			"msg_id":        msg.ID,
			logrus.ErrorKey: err,
		}).Debug("failed to flush buffered writer")
		return s.handleErr(err)
	}

	if err = w.Close(); err != nil {
		s.log.WithFields(logrus.Fields{
			"provider_id":   s.p.UID,
			"msg_id":        msg.ID,
			logrus.ErrorKey: err,
		}).Debug("failed to close data writer")
		return s.handleErr(err)
	}

	if err = c.Quit(); err != nil {
		s.log.WithFields(logrus.Fields{
			"provider_id":   s.p.UID,
			"msg_id":        msg.ID,
			logrus.ErrorKey: err,
		}).Debug("failed to quit smtp client connection")
		return s.handleErr(err)
	}
	return nil
}

func (s *SMTPProvider) handleErr(err error) error {
	var protoErr *textproto.Error
	if errors.As(err, &protoErr) {
		switch {
		case protoErr.Code == 535:
			// The 535 means Authentication credentials invalid.
			return mailprovider2.ErrAuth(err)
		case protoErr.Code >= 400 && protoErr.Code < 500:
			// The code between 400 and 500 is a temporary error.
			return mailprovider2.ErrTemporary(err)
		case protoErr.Code >= 500 && protoErr.Code < 600:
			// The code between 500 and 600 is a permanent error.
			return mailprovider2.ErrPermanent(err)
		}
	}
	return err
}

// writeMessage writes the message to the buffer.
func (s *SMTPProvider) writeMessage(msg *message.Message, buf *bufio.Writer) error {
	// Write the content type.
	s.writeLine(buf, "MIME-version: 1.0")
	s.writeHeaderLine(buf, "Content-Type", msg.ContentType)
	s.writeHeaderLine(buf, "From", msg.From.String())
	var sb strings.Builder
	for _, to := range msg.To {
		sb.WriteString(to.String())
		sb.WriteString(";")
	}
	s.writeHeaderLine(buf, "To", sb.String())
	s.writeHeaderLine(buf, "Subject", mime.QEncoding.Encode("UTF-8", msg.Subject))
	s.writeHeaderLine(buf, "Content-Transfer-Encoding", "base64")
	buf.WriteString("\r\n")

	temp := make([]byte, base64.StdEncoding.EncodedLen(len(msg.Body)))
	base64.StdEncoding.Encode(temp, []byte(msg.Body))
	buf.Write(temp)

	return nil
}

func (s *SMTPProvider) writeLine(buf *bufio.Writer, line string) {
	buf.WriteString(line)
	buf.WriteString("\r\n")
}

func (s *SMTPProvider) writeHeaderLine(buf *bufio.Writer, header, content string) {
	buf.WriteString(header)
	buf.WriteString(": ")
	buf.WriteString(content)
	buf.WriteString("\r\n")
}

func (s *SMTPProvider) addr() string {
	s.l.RLock()
	defer s.l.RUnlock()

	return fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
}

func (s *SMTPProvider) auth() smtp.Auth {
	s.l.RLock()
	defer s.l.RUnlock()

	return smtp.PlainAuth("", s.cfg.Username.UnsafeString(), s.cfg.Password.UnsafeString(), s.cfg.Host.UnsafeString())
}
