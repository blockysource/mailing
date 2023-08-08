package emailhandler

import (
	mailprovidermanager "github.com/blockysource/mailing/logic/mailprovider/manager"
)

// Handler is a handler that handles emails.
type Handler struct {
	m *mailprovidermanager.Manager
}
