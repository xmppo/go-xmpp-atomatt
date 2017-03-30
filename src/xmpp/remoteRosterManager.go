package xmpp

import (
	"encoding/xml"
)

const (
	NSRemoteRosterManager = "urn:xmpp:tmp:roster-management:0"

	RemoteRosterManagerTypeRequest  = "request"
	RemoteRosterManagerTypeAllowed  = "allowed"
	RemoteRosterManagerTypeRejected = "rejected"
)

// XEP-0321: Remote Roster Manager

type RemoteRosterManagerQuery struct {
	XMLName xml.Name `xml:"urn:xmpp:tmp:roster-management:0 query"`
	Reason  string   `xml:"reason,attr,omitempty"`
	Type    string   `xml:"type,attr"`
}
