package main

import (
	"flag"
	"log"
	"xmpp"
)

func main() {

	jid := flag.String("jid", "", "JID")
	secret := flag.String("secret", "", "Component secret")
	flag.Parse()

	jid2, err := xmpp.ParseJID(*jid)
	if err != nil {
		log.Fatal(err)
	}

	c, err := xmpp.NewComponent("localhost:5347", jid2, *secret)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(c)
	select {}
}
