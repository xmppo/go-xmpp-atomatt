package main

import (
	"flag"
	"log"
	"strconv"
	"time"
	"xmpp"
)

var _ = log.Println

func main() {

	flag.Parse()

	switch flag.Arg(0) {
	case "producer":
		producer(flag.Args()[1:])
	case "consumer":
		consumer(flag.Args()[1:])
	default:
		log.Fatal("producer or consumer?")
	}
}

func producer(args []string) {

	flags := flag.NewFlagSet("producer", flag.ExitOnError)
	jidFlag := flags.String("jid", "", "JID")
	passFlag := flags.String("pass", "", "Password")
	insecureFlag := flags.Bool("insecure", false, "Allow insecure TLS")
	toFlag := flags.String("to", "", "Recipient")
	countFlag := flags.Int("count", 1, "Number of messages to send")
	flags.Parse(args)

	jid := must(xmpp.ParseJID(*jidFlag)).(xmpp.JID)
	addrs := must(xmpp.HomeServerAddrs(jid)).([]string)
	stream := must(xmpp.NewStream(addrs[0], nil)).(*xmpp.Stream)
	config := xmpp.ClientConfig{InsecureSkipVerify: *insecureFlag}
	x := must(xmpp.NewClientXMPP(stream, jid, *passFlag, &config)).(*xmpp.XMPP)

	x.Out <- xmpp.Presence{}

	go func() {
		count := *countFlag
		for i := 0; i < count; i++ {
			x.Out <- xmpp.Message{From: jid.String(), To: *toFlag, Body: strconv.Itoa(i)}
		}
		close(x.Out)
	}()

	for stanza := range x.In {
		switch v := stanza.(type) {
		case error:
			log.Fatal(v)
		default:
			log.Println(stanza)
		}
	}
}

func consumer(args []string) {

	flags := flag.NewFlagSet("consumer", flag.ExitOnError)
	jidFlag := flags.String("jid", "", "JID")
	passFlag := flags.String("pass", "", "Password")
	insecureFlag := flags.Bool("insecure", false, "Allow insecure TLS")
	serverFlag := flags.String("server", "", "XMPP server address")
	flags.Parse(args)

	jid := must(xmpp.ParseJID(*jidFlag)).(xmpp.JID)
	var x *xmpp.XMPP
	if jid.Node == "" {
		stream := must(xmpp.NewStream(*serverFlag, nil)).(*xmpp.Stream)
		x = must(xmpp.NewComponentXMPP(stream, jid, *passFlag)).(*xmpp.XMPP)
	} else {
		addrs := must(xmpp.HomeServerAddrs(jid)).([]string)
		stream := must(xmpp.NewStream(addrs[0], nil)).(*xmpp.Stream)
		config := xmpp.ClientConfig{InsecureSkipVerify: *insecureFlag}
		x = must(xmpp.NewClientXMPP(stream, jid, *passFlag, &config)).(*xmpp.XMPP)
		x.Out <- xmpp.Presence{}
	}

	count := 0
	throughputCount := 0

	go func() {
		throughput := time.Tick(time.Second)
		total := time.Tick(time.Second * 5)
		for {
			select {
			case <-throughput:
				log.Printf("throughput: %d msgs/s\n", count-throughputCount)
				throughputCount = count
			case <-total:
				log.Printf("total: %d\n", count)
			}
		}
	}()

	for stanza := range x.In {
		switch v := stanza.(type) {
		case *xmpp.Message:
			count++
		case error:
			log.Fatal(v)
		}
	}
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		log.Fatal(err)
	}
	return v
}
