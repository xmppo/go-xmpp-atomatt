package main

import (
	"flag"
	"log"
	"xmpp"
)

var (
	addr       = ""
	skipverify = false
	jid        = ""
	pass       = ""
)

func init() {
	flag.StringVar(&addr, "addr", "", "XMPP server address, <host>:<port>. Optional")
	flag.BoolVar(&skipverify, "skipverify", false, "Skip TLS certificate verification.")
	flag.StringVar(&jid, "jid", "", "User's JID, e.g. alice@wonderland.lit/chat.")
	flag.StringVar(&pass, "pass", "", "User's password.")
}

func main() {

	// Parse args
	flag.Parse()
	jid, _ := xmpp.ParseJID(jid)

	// Lookup home server TCP addr if not explicitly set.
	if addr == "" {
		if addrs, err := xmpp.HomeServerAddrs(jid); err != nil {
			log.Fatal(err)
		} else {
			addr = addrs[0]
		}
	}

	// Create stream.
	stream, err := xmpp.NewStream(addr)
	if err != nil {
		log.Fatal(err)
	}

	// Configure stream as client.
	config := xmpp.ClientConfig{InsecureSkipVerify: skipverify}
	x, err := xmpp.NewClientXMPP(stream, jid, pass, &config)
	if err != nil {
		log.Fatal(err)
	}

	// Signal presence.
	x.Send(xmpp.Presence{})

	// Log anything that arrives.
	for {
		stanza, err := x.Recv()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("recv: %T %v", stanza, stanza)
	}
}
