package xmpp

import (
	"encoding/xml"
	"strings"
)

const (
	nsDiscoInfo  = "http://jabber.org/protocol/disco#info"
	nsDiscoItems = "http://jabber.org/protocol/disco#items"
)

// Service Discovery (XEP-0030) protocol. "Wraps" XMPP instance to provide a
// more convenient API for Disco clients.
type Disco struct {
	XMPP *XMPP
}

// Iq get/result payload for "info" requests.
type DiscoInfo struct {
	XMLName  xml.Name        `xml:"http://jabber.org/protocol/disco#info query"`
	Identity []DiscoIdentity `xml:"identity"`
	Feature  []DiscoFeature  `xml:"feature"`
}

// Identity
type DiscoIdentity struct {
	Category string `xml:"category,attr"`
	Type     string `xml:"type,attr"`
	Name     string `xml:"name,attr"`
}

// Feature
type DiscoFeature struct {
	Var string `xml:"var,attr"`
}

// Iq get/result payload for "items" requests.
type DiscoItems struct {
	XMLName xml.Name    `xml:"http://jabber.org/protocol/disco#items query"`
	Item    []DiscoItem `xml:"item"`
}

// Item.
type DiscoItem struct {
	JID  string `xml:"jid,attr"`
	Node string `xml:"node,attr"`
	Name string `xml:"name,attr"`
}

// Request information about the service identified by 'to'.
func (disco *Disco) Info(to string, from string) (*DiscoInfo, error) {

	if from == "" {
		from = disco.XMPP.JID.Full()
	}

	req := &Iq{Id: UUID4(), Type: "get", To: to, From: from}
	req.PayloadEncode(&DiscoInfo{})

	resp, err := disco.XMPP.SendRecv(req)
	if err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, resp.Error
	}

	info := &DiscoInfo{}
	resp.PayloadDecode(info)

	return info, err
}

// Request items in the service identified by 'to'.
func (disco *Disco) Items(to string, from string) (*DiscoItems, error) {

	if from == "" {
		from = disco.XMPP.JID.Full()
	}

	req := &Iq{Id: UUID4(), Type: "get", To: to, From: from}
	req.PayloadEncode(&DiscoItems{})

	resp, err := disco.XMPP.SendRecv(req)
	if err != nil {
		return nil, err
	} else if resp.Error != nil {
		return nil, resp.Error
	}

	items := &DiscoItems{}
	resp.PayloadDecode(items)

	return items, err
}

var discoNamespacePrefix = strings.Split(nsDiscoInfo, "#")[0]

// Matcher instance to match <iq/> stanzas with a disco payload.
var DiscoPayloadMatcher = MatcherFunc(
	func(v interface{}) bool {
		iq, ok := v.(*Iq)
		if !ok {
			return false
		}
		ns := strings.Split(iq.PayloadName().Space, "#")[0]
		return ns == discoNamespacePrefix
	},
)
