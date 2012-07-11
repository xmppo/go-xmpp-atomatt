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

	// Parse args.
	flag.Parse()
	jid, _ := xmpp.ParseJID(*jid)
	password := *password

	// Create stream.
	stream, err := xmpp.NewStream(jid.Domain + ":5222", &xmpp.StreamConfig{LogStanzas: true})
	if err != nil {
		log.Fatal(err)
	}

	// Configure stream as a client connection.
	x, err := xmpp.NewClientXMPP(stream, jid, password, &xmpp.ClientConfig{InsecureSkipVerify: true})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connection established for %s\n", x.JID)

	// Announce presence.
	x.Send(xmpp.Presence{})

	// Filter messages into dedicated channel and start a thread to log them.
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
	iq := xmpp.Iq{Id: "abc", Type: "get", To: x.JID.Domain}
	iq.PayloadEncode(info)
	reply, _ := x.SendRecv(&iq)
	reply.PayloadDecode(info)
	log.Printf("* info: %v\n", info)

	select {}
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
