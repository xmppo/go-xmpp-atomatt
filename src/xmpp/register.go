package xmpp

import (
	"encoding/xml"
)

const (
	NSRegister = "jabber:iq:register"
)

// XEP-0077: In-Band Registration

type RegisterQuery struct {
	XMLName      xml.Name            `xml:"jabber:iq:register query"`
	Instructions string              `xml:"instructions"`
	Username     string              `xml:"username"`
	Password     string              `xml:"password"`
	XForm        AdHocXForm          `xml:"x"`
	Registered   *RegisterRegistered `xmp:"registered"`
	Remove       *RegisterRemove     `xmp:"remove"`
}

type RegisterRegistered struct {
	XMLName xml.Name `xml:"registered"`
}

type RegisterRemove struct {
	XMLName xml.Name `xml:"remove"`
}
