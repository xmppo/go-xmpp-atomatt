package main

import (
	"flag"
	"log"
	"xmpp"
)

var (
	addr   = flag.String("a", "", "Server component address")
	jid    = flag.String("j", "", "JID")
	secret = flag.String("s", "", "Component secret")
)

func main() {

	flag.Parse()

	// Create stream and configure it as a component connection.
	jid := must(xmpp.ParseJID(*jid)).(xmpp.JID)
	stream := must(xmpp.NewStream(*addr, &xmpp.StreamConfig{LogStanzas: true})).(*xmpp.Stream)
	comp := must(xmpp.NewComponentXMPP(stream, jid, *secret)).(*xmpp.XMPP)

	for x := range comp.In {
		log.Printf("recv: %v", x)
	}
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		log.Fatal(err)
	}
	return v
}
