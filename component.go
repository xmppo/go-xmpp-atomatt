package main

import (
	"flag"
	"log"
	"xmpp"
)

var (
	jid = flag.String("j", "", "JID")
	secret = flag.String("s", "", "Component secret")
)

func main() {
	flag.Parse()
	jid, _ := xmpp.ParseJID(*jid)
	secret := *secret

	x, err := xmpp.NewComponentXMPP("localhost:5347", jid, secret)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(x)
	select {}
}
