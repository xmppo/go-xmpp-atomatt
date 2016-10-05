package xmpp

import (
	"encoding/xml"
)

const (
  NSPing = "urn:xmpp:ping"
)

type Ping struct {
	XMLName xml.Name           `xml:"urn:xmpp:ping ping"`
}
