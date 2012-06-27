package main

import (
	"flag"
	"log"
	"xmpp"
)

func main() {

	jid := flag.String("jid", "", "JID")
	password := flag.String("pass", "", "Password")
	flag.Parse()

	jid2, err := xmpp.ParseJID(*jid)
	if err != nil {
		log.Fatal(err)
	}

	c, err := xmpp.NewClient(jid2, *password, &xmpp.ClientConfig{})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(c)
	select {}
}
