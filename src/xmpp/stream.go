package xmpp

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net"
)

const (
	nsStream = "http://etherx.jabber.org/streams"
	nsTLS = "urn:ietf:params:xml:ns:xmpp-tls"
)

type Stream struct {
	conn net.Conn
	dec *xml.Decoder
}

func NewStream(addr string) (*Stream, error) {

	log.Println("Connecting to", addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	dec := xml.NewDecoder(conn)
	return &Stream{conn, dec}, nil
}

func (stream *Stream) UpgradeTLS(config *tls.Config) error {

	log.Println("Upgrading to TLS")

	if err := stream.Send("<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>"); err != nil {
		return err
	}

	p := tlsProceed{}
	if err := stream.Decode(&p); err != nil {
		return err
	}

	conn := tls.Client(stream.conn, &tls.Config{InsecureSkipVerify: true})
	if err := conn.Handshake(); err != nil {
		return err
	}

	stream.conn = conn
	stream.dec = xml.NewDecoder(stream.conn)

	return nil
}

func (stream *Stream) Send(s string) error {
	if _, err := stream.conn.Write([]byte(s)); err != nil {
		return err
	}
	return nil
}

func (stream *Stream) Next(match *xml.Name) (*xml.StartElement, error) {
	for {
		t, err := stream.dec.Token()
		if err != nil {
			return nil, err
		}
		if e, ok := t.(xml.StartElement); ok {
			if match != nil && e.Name != *match {
				return nil, errors.New(fmt.Sprintf("Expected %s, got %s", *match, e.Name))
			}
			return &e, nil
		}
	}
	panic("Unreachable")
}

func (stream *Stream) Decode(i interface{}) error {
	return stream.dec.Decode(i)
}

func (stream *Stream) DecodeElement(i interface{}, se *xml.StartElement) error {
	return stream.dec.DecodeElement(i, se)
}

type tlsProceed struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls proceed"`
}
