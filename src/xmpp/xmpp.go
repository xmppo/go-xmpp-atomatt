package xmpp

import (
	"fmt"
	"log"
	"sync"
)

// Handles XMPP conversations over a Stream. Use NewClientXMPP and/or
// NewComponentXMPP to create and configuring a XMPP instance.
type XMPP struct {
	// JID associated with the stream. Note: this may be negotiated with the
	// server during setup and so must be used for all messages.
	JID JID
	stream *Stream

	// Stanza channels.
	in chan interface{}
	out chan interface{}

	// Incoming stanza filters.
	filterLock sync.Mutex
	nextFilterId FilterId
	filters map[FilterId]filter
}

func newXMPP(jid JID, stream *Stream) *XMPP {
	x := &XMPP{
		JID: jid,
		stream: stream,
		in: make(chan interface{}),
		out: make(chan interface{}),
		filters: make(map[FilterId]filter),
	}
	go x.sender()
	go x.receiver()
	return x
}

// Send a stanza.
func (x *XMPP) Send(v interface{}) {
	x.out <- v
}

// Return the next stanza.
func (x *XMPP) Recv() (interface{}, error) {
	v := <-x.in
	if err, ok := v.(error); ok {
		return nil, err
	}
	return v, nil
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

	// Protect against concurrent access.
	x.filterLock.Lock()
	defer x.filterLock.Lock()

	// Create filter chan and add to map.
	filterId := x.nextFilterId
	x.nextFilterId ++
	ch := make(chan interface{})
	x.filters[filterId] = filter{fn, ch}

	return filterId, ch
}

func (x *XMPP) RemoveFilter(id FilterId) error {

	// Protect against concurrent access.
	x.filterLock.Lock()
	defer x.filterLock.Lock()

	// Find filter.
	filter, ok := x.filters[id]
	if !ok {
		return id
	}

	// Close filter channel and remove from map.
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

	defer close(x.in)

	for {
		start, err := x.stream.Next(nil)
		if err != nil {
			x.in <- err
			return
		}

		var v interface{}
		switch start.Name.Local {
		case "error":
			v = &Error{}
		case "iq":
			v = &Iq{}
		case "message":
			v = &Message{}
		case "presence":
			v = &Presence{}
		default:
			log.Fatal("Unexected element: %T %v", start, start)
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
