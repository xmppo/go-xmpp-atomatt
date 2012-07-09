package xmpp

import (
	"fmt"
	"log"
)

// Handles XMPP conversations over a Stream. Use NewClientXMPP and/or
// NewComponentXMPP to create and configuring a XMPP instance.
type XMPP struct {
	// JID associated with the stream. Note: this may be negotiated with the
	// server during setup and so must be used for all messages.
	JID JID
	stream *Stream
	in chan interface{}
	out chan interface{}
	nextFilterId FilterId
	filters map[FilterId]filter
}

func newXMPP(jid JID, stream *Stream) *XMPP {
	x := &XMPP{
		jid,
		stream,
		make(chan interface{}),
		make(chan interface{}),
		0,
		make(map[FilterId]filter),
	}
	go x.sender()
	go x.receiver()
	return x
}

// Send a stanza.
func (x *XMPP) Send(v interface{}) {
	x.out <- v
}

func (x *XMPP) Recv() interface{} {
	return <-x.in
}

func (x *XMPP) SendRecv(iq *Iq) (*Iq, error) {

	fid, ch := x.AddFilter(IqResult(iq.Id))
	defer x.RemoveFilter(fid)

	x.Send(iq)

	stanza := <-ch
	reply, ok := stanza.(*Iq)
	if !ok {
		return nil, fmt.Errorf("Expected Iq, for %T", stanza)
	}
	return reply, nil
}

type FilterId int64

func (fid FilterId) Error() string {
	return fmt.Sprintf("Invalid filter id: %d", fid)
}

func (x *XMPP) AddFilter(fn FilterFn) (FilterId, chan interface{}) {
	ch := make(chan interface{})
	filterId := x.nextFilterId
	x.nextFilterId ++
	x.filters[filterId] = filter{fn, ch}
	return filterId, ch
}

func (x *XMPP) RemoveFilter(id FilterId) error {
	filter, ok := x.filters[id]
	if !ok {
		return id
	}
	close(filter.ch)
	delete(x.filters, id)
	return nil
}

func IqResult(id string) FilterFn {
	return func(v interface{}) bool {
		iq, ok := v.(*Iq)
		if !ok {
			return false
		}
		if iq.Id != id {
			return false
		}
		return true
	}
}

type FilterFn func(v interface{}) bool

type filter struct {
	fn FilterFn
	ch chan interface{}
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

		filtered := false
		for _, filter := range x.filters {
			if filter.fn(v) {
				filter.ch <- v
				filtered = true
			}
		}

		if !filtered {
			x.in <- v
		}
	}
}

// BUG(matt): filter id generation is not re-entrant.

// BUG(matt): filters map is not re-entrant.
