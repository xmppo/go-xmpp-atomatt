package main

import (
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

	x, err := xmpp.NewClientXMPP(jid, password, &xmpp.ClientConfig{InsecureSkipVerify: true})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connection established for %s\n", x.JID)
	x.Send(xmpp.Presence{})

	select {}
}
