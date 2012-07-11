package main

import (
	"encoding/xml"
	"flag"
	"log"
	"xmpp"
)

var (
	jid      = flag.String("j", "", "JID")
	password = flag.String("p", "", "Password")
)

func main() {

	flag.Parse()

	// Create stream and configure it as a client connection.
	jid := must(xmpp.ParseJID(*jid)).(xmpp.JID)
	stream := must(xmpp.NewStream(jid.Domain + ":5222", &xmpp.StreamConfig{LogStanzas: true})).(*xmpp.Stream)
	x := must(xmpp.NewClientXMPP(stream, jid, *password, &xmpp.ClientConfig{InsecureSkipVerify: true})).(*xmpp.XMPP)

	log.Printf("Connection established for %s\n", x.JID)

	// Announce presence.
	x.Send(xmpp.Presence{})

	// Filter messages into dedicated channel and start a goroutine to log them.
	_, messages := x.AddFilter(
		func(v interface{}) bool {
			_, ok := v.(*xmpp.Message)
			return ok
		},
	)
	go func() {
		for message := range messages {
			log.Printf("* message: %v\n", message)
		}
	}()

	// Log any stanzas that are not handled elsewhere.
	go func() {
		for {
			stanza, err := x.Recv()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("* recv: %v\n", stanza)
		}
	}()

	// Get disco#info for home server.
	info := &DiscoInfo{}
	iq := xmpp.Iq{Id: xmpp.UUID4(), Type: "get", To: x.JID.Domain}
	iq.PayloadEncode(info)
	reply, _ := x.SendRecv(&iq)
	reply.PayloadDecode(info)
	log.Printf("* info: %v\n", info)

	select {}
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

type DiscoInfo struct {
	XMLName  xml.Name        `xml:"http://jabber.org/protocol/disco#info query"`
	Identity []DiscoIdentity `xml:"identity"`
	Feature  []DiscoFeature  `xml:"feature"`
}

type DiscoIdentity struct {
	Type     string `xml:"type,attr"`
	Name     string `xml:"name,attr"`
	Category string `xml:"category,attr"`
}

type DiscoFeature struct {
	Var string `xml:"var,attr"`
}
