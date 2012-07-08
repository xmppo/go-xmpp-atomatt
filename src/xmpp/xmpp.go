package xmpp

import "log"

type XMPP struct {
	JID JID
	stream *Stream
	in chan interface{}
	out chan interface{}
}

func newXMPP(jid JID, stream *Stream) *XMPP {
	x := &XMPP{
		jid,
		stream,
		make(chan interface{}),
		make(chan interface{}),
	}
	go x.sender()
	go x.receiver()
	return x
}

func (x *XMPP) Send(v interface{}) {
	x.out <- v
}

func (x *XMPP) sender() {
	for v := range x.out {
		x.stream.Send(v)
	}
}

func (x *XMPP) receiver() {
	for {
		start, err := x.stream.Next(nil)
		if err != nil {
			log.Fatal(err)
		}

		var v interface{}
		switch start.Name.Local {
		case "iq":
			v = &Iq{}
		case "message":
			v = &Message{}
		case "presence":
			v = &Presence{}
		default:
			panic("Unexected element: " + start.Name.Local)
		}

		err = x.stream.DecodeElement(v, start)
		if err != nil {
			log.Fatal(err)
		}

		x.in <- v
	}
}
