package xmpp

import (
	"encoding/xml"
)

const (
	NSRoster = "jabber:iq:roster"

	RosterSubscriptionBoth   = "both"
	RosterSubscriptionFrom   = "from"
	RosterSubscriptionTo     = "to"
	RosterSubscriptionRemove = "remove"
)

type RosterQuery struct {
	XMLName xml.Name     `xml:"jabber:iq:roster query"`
	Items   []RosterItem `xml:"item"`
}

type RosterItem struct {
	JID          string   `xml:"jid,attr"`
	Name         string   `xml:"name,attr,omitempty"`
	Subscription string   `xml:"subscription,attr"`
	Groupes      []string `xml:"group"`
}
