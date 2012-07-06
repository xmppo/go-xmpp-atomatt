package xmpp

import (
	"fmt"
	"strings"
)

/*
Jabber Identifier - uniquely identifies an individual entity in a XMPP/Jabber
network.
*/
type JID struct {
	// Node/local component e.g. the alice of alice@example.com/foo.
	Node string

	// Domain component, e.g. the example.com of alice@example.com/foo for a
	// client or the whole JID of a component.
	Domain string

	// Resource component, e.g. the foo of alice@example.com/foo.
	Resource string
}

// Return the "bare" JID, i.e. no resource component.
func (jid JID) Bare() string {
	if jid.Node == "" {
		return jid.Domain
	}
	return fmt.Sprintf("%s@%s", jid.Node, jid.Domain)
}

// Return the full JID as a string.
func (jid JID) Full() string {
	if jid.Resource == "" {
		return jid.Bare()
	}
	return fmt.Sprintf("%s@%s/%s", jid.Node, jid.Domain, jid.Resource)
}

// Return full JID as a string.
func (jid JID) String() string {
	return jid.Full()
}

// Parse a string into a JID structure.
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
		jid.Node = parts[0]
		jid.Domain = parts[1]
	}

	return
}

// BUG(matt): ParseJID should fail for incorrectly formatted JIDs.
