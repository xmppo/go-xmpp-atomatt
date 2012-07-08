package xmpp

import (
	"bytes"
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

// Create a XML stream connection. Typically handles the encoding and decoding
// of XML data for a higher-level API, e.g. XMPP.
func NewStream(addr string) (*Stream, error) {

	log.Println("Connecting to", addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	if _, err := conn.Write([]byte("<?xml version='1.0'?>")); err != nil {
		return nil, err
	}

	dec := xml.NewDecoder(conn)
	return &Stream{conn, dec}, nil
}

// Upgrade the stream's underlying net conncetion to TLS.
func (stream *Stream) UpgradeTLS(config *tls.Config) error {

	conn := tls.Client(stream.conn, config)
	if err := conn.Handshake(); err != nil {
		return err
	}

	stream.conn = conn
	stream.dec = xml.NewDecoder(stream.conn)

	return nil
}

// Send the element's start tag. Typically used to open the stream's document.
func (stream *Stream) SendStart(start *xml.StartElement) error {

	buf := new(bytes.Buffer)
	if _, err := buf.Write([]byte{'<'}); err != nil {
		return err
	}
	if err := writeXMLName(buf, start.Name); err != nil {
		return err
	}
	for _, attr := range start.Attr {
		if _, err := buf.Write([]byte{' '}); err != nil {
			return err
		}
		if err := writeXMLAttr(buf, attr); err != nil {
			return err
		}
	}
	if _, err := buf.Write([]byte{'>'}); err != nil {
		return err
	}

	return stream.send(buf.Bytes())
}

// Send a stanza. Used to write a complete, top-level element.
func (stream *Stream) Send(v interface{}) error {
	bytes, err := xml.Marshal(v)
	if err != nil {
		return err
	}
	return stream.send(bytes)
}

func (stream *Stream) send(b []byte) error {
	log.Println("send:", string(b))
	if _, err := stream.conn.Write(b); err != nil {
		return err
	}
	return nil
}

// Find start of next stanza. If match is not nil the stanza's XML name
// compared and must be equal.
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

// Decode the next stanza. Works like xml.Unmarshal but reads from the stream's
// connection.
func (stream *Stream) Decode(v interface{}) error {
	return stream.dec.Decode(v)
}

// Decode the stanza with the given start element. Works like
// xml.Decoder.DecodeElement.
func (stream *Stream) DecodeElement(v interface{}, start *xml.StartElement) error {
	return stream.dec.DecodeElement(v, start)
}
