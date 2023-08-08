package mailprovider

import (
	"net/mail"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	mailingadminv1 "github.com/blockysource/go-genproto/blockyapis/mailing/admin/v1alpha"
)

// Base is a base structure used to provide common provider fields.
type Base struct {
	ID          string
	Type        mailingadminv1.MailingProviderType
	Name        string
	FromAddress *mail.Address
	IsVerified  bool
}

// MailingProviderDefinition is a mailing provider.
type MailingProviderDefinition struct {
	// UID is the unique identifier of the mailing provider.
	UID string
	// CreatedAt is the creation time of the mailing provider.
	CreatedAt time.Time
	// UpdatedAt is the update time of the mailing provider.
	UpdatedAt time.Time
	// FromAddress is the default from address of the mailing provider.
	FromAddress string
	// Name is the name of the mailing provider.
	Name string
	// Type is the type of the mailing provider.
	Type mailingadminv1.MailingProviderType
	// Config is the configuration of the mailing provider.
	Config *mailingadminv1.MailingProviderConfig
	// InUse is a flag that indicates whether the mailing provider is in use.
	InUse bool
	// VerifiedAt is the verification time of the mailing provider.
	VerifiedAt time.Time
}

// ToProto converts the mailing provider to a protobuf message.
func (d MailingProviderDefinition) ToProto() mailingadminv1.MailingProvider {
	var verifiedAt *timestamppb.Timestamp
	if !d.VerifiedAt.IsZero() {
		verifiedAt = timestamppb.New(d.VerifiedAt)
	}
	return mailingadminv1.MailingProvider{
		Uid:         d.UID,
		CreatedAt:   timestamppb.New(d.CreatedAt),
		UpdatedAt:   timestamppb.New(d.UpdatedAt),
		FromAddress: d.FromAddress,
		Name:        d.Name,
		Type:        d.Type,
		Config:      proto.Clone(d.Config).(*mailingadminv1.MailingProviderConfig),
		InUse:       d.InUse,
		VerifiedAt:  verifiedAt,
	}
}
