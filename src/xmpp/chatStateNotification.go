package xmpp

import (
	"encoding/xml"
)

const (
	NSChatStatesNotification = "http://jabber.org/protocol/chatstates"
)

// XEP-0085: Chat States Notification

type Active struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates active"`
}
type Composing struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates composing"`
}
type Paused struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates paused"`
}
type Inactive struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates inactive"`
}
type Gone struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates gone"`
}
