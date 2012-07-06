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

	stream, err := xmpp.ClientStream(jid, password, &xmpp.ClientConfig{})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(stream)
	select {}
}
