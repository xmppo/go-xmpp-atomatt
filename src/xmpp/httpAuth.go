package xmpp

import (
	"encoding/xml"
)

const (
	NSHTTPAuth = "http://jabber.org/protocol/http-auth"
)

// XEP-0070: Verifying HTTP Requests via XMPP
type Confirm struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/http-auth confirm"`
	Id      string   `xml:"id,attr"`
	Method  string   `xml:"method,attr"`
	URL     string   `xml:"url,attr"`
}
