package xmpp

import (
	"encoding/xml"
)

const (
	NsVCardTemp = "vcard-temp"
)

// XEP-0054 vCard
type VCard struct {
	XMLName xml.Name `xml:"vcard-temp vCard"`
	// TODO Must complete truct
}
