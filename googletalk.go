package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"xmpp"
)

var (
	jid = flag.String("jid", "", "JID")
	pass = flag.String("pass", "", "Password")
)

func main() {

	flag.Parse()
	jid, _ := xmpp.ParseJID(*jid)
	pass := *pass

	// Lookup XMPP client net addr.
	addr, err := xmppHomeAddr(jid)
	if err != nil {
		log.Fatal(err)
	}

	// Create stream.
	stream, err := xmpp.NewStream(addr)
	if err != nil {
		log.Fatal(err)
	}

	// Configure stream as client.
	x, err := xmpp.NewClientXMPP(stream, jid, pass, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Signal presence.
	x.Send(xmpp.Presence{})

	for {
		stanza, err := x.Recv()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("recv: %v", stanza)
	}
}

// Get XMPP server addr from DNS SRV entry.
func xmppHomeAddr(jid xmpp.JID) (addr string, err error) {

	_, addrs, err := net.LookupSRV("xmpp-client", "tcp", jid.Domain)
	if err != nil {
		return
	}

	if len(addrs) == 0 {
		err = fmt.Errorf("No addrs for %s", jid.Domain)
		return
	}

	return fmt.Sprintf("%s:%d", addrs[0].Target, addrs[0].Port), nil
}
