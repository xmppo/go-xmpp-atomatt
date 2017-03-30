package xmpp

import (
	"encoding/xml"
)

const (
	NSJabberClient = "jabber:iq:version"
)

// XEP-0092 Software Version
type SoftwareVersion struct {
	XMLName xml.Name `xml:"jabber:iq:version query"`
	Name    string   `xml:"name,omitempty"`
	Version string   `xml:"version,omitempty"`
	OS      string   `xml:"os,omitempty"`
}
