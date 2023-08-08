package smtpmailprovider

import (
	"net"

	"github.com/sirupsen/logrus"

	"github.com/blockysource/blocky/pkg/go/geoip"
	"github.com/blockysource/blocky/services/mailing/public/mailing"
)

// SMTPProvidersConfig is a configuration used for SMTP providers.
// It provides common configuration for the SMTP providers, like the domain name used for the HELO command.
type SMTPProvidersConfig struct {
	Domain string
}

// NewSMTPProvidersConfig creates a new SMTP providers configuration.
func NewSMTPProvidersConfig(cfg *mailing.Config, log *logrus.Entry) (*SMTPProvidersConfig, error) {
	c, err := geoip.DefaultConsensus(geoip.DefaultConsensusConfig(), log)
	if err != nil {
		return nil, err
	}

	ip, err := c.ExternalIP()
	if err != nil {
		return nil, err
	}

	names, err := net.LookupAddr(ip.String())
	if err != nil {
		return nil, err
	}

	domain := cfg.Domain
	if cfg.Domain == "" {
		if len(names) == 0 {
			domain = ip.String()
			log.WithField("ip", ip).
				Warn("no domain name found for IP address, using IP address as domain name")
		} else {
			domain = names[0]
			log.WithFields(logrus.Fields{
				"ip":     ip,
				"domain": domain,
			}).Info("using domain name for IP address")
		}
	} else {
		// Check if any of the domain names found for the IP address matches the configured domain name.
		found := false
		for _, name := range names {
			if name == domain {
				found = true
				break
			}
		}
		if !found {
			log.WithFields(logrus.Fields{
				"ip":       ip,
				"domain":   domain,
				"resolved": names,
			}).Warn("configured domain name does not match any of the domain names found for the IP address")
		}
	}

	return &SMTPProvidersConfig{Domain: domain}, nil
}
