package main

import (
	"flag"
	"log"
	"xmpp"
)

var (
	addr = flag.String("a", "", "Server component address")
	jid = flag.String("j", "", "JID")
	secret = flag.String("s", "", "Component secret")
)

func main() {
	flag.Parse()
	addr := *addr
	jid, _ := xmpp.ParseJID(*jid)
	secret := *secret

	// Create stream.
	stream, err := xmpp.NewStream(addr)
	if err != nil {
		log.Fatal(err)
	}

	// Configure stream as a component connection.
	x, err := xmpp.NewComponentXMPP(stream, jid, secret)
	if err != nil {
		log.Fatal(err)
	}

	for {
		v := x.Recv()
		log.Printf("recv: %v", v)
	}
}
