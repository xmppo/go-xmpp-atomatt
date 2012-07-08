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

	// Create stream.
	stream, err := xmpp.NewStream(jid.Domain + ":5222")
	if err != nil {
		log.Fatal(err)
	}

	// Configure stream as a client connection.
	x, err := xmpp.NewClientXMPP(stream, jid, password, &xmpp.ClientConfig{InsecureSkipVerify: true})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connection established for %s\n", x.JID)
	x.Send(xmpp.Presence{})
	x.Send(xmpp.Message{To: "carol@localhost", Body: "Hello!"})

	select {}
}
