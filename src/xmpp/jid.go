package xmpp

import (
	"fmt"
	"strings"
)

type JID struct {
	Local string
	Domain string
	Resource string
}

func (jid JID) Bare() string {
	if jid.Local == "" {
		return jid.Domain
	}
	return fmt.Sprintf("%s@%s", jid.Local, jid.Domain)
}

func (jid JID) String() string {
	if jid.Resource == "" {
		return jid.Bare()
	}
	return fmt.Sprintf("%s@%s/%s", jid.Local, jid.Domain, jid.Resource)
}

func ParseJID(s string) (jid JID, err error) {

	if parts := strings.SplitN(s, "/", 2); len(parts) == 1 {
		s = parts[0]
	} else {
		s = parts[0]
		jid.Resource = parts[1]
	}

	if parts := strings.SplitN(s, "@", 2); len(parts) != 2 {
		jid.Domain = parts[0]
	} else {
		jid.Local = parts[0]
		jid.Domain = parts[1]
	}

	return
}
