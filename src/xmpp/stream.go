package xmpp

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"io"
	"log"
	"net"
)

// Stream configuration.
type StreamConfig struct {
	// Log all sent and received stanzas.
	// Enabling this option causes stanzas to be buffered in memory before they
	// are either sent to the server or delivered to the application. It also
	// causes incoming stanzas to be XML-parsed a second time.
	LogStanzas bool
}

type Stream struct {
	conn              net.Conn
	dec               *xml.Decoder
	config            *StreamConfig
	stanzaBuf         string
	incomingNamespace nsMap
}

// Create a XML stream connection. A Steam is used by an XMPP instance to
// handle sending and receiving XML data over the net connection.
func NewStream(addr string, config *StreamConfig) (*Stream, error) {

	if config == nil {
		config = &StreamConfig{}
	}

	log.Println("Connecting to", addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	stream := &Stream{conn: conn, dec: xml.NewDecoder(conn), config: config}

	if err := stream.send([]byte("<?xml version='1.0' encoding='utf-8'?>")); err != nil {
		return nil, err
	}

	return stream, nil
}

// Upgrade the stream's underlying net connection to TLS.
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
func (stream *Stream) SendStart(start *xml.StartElement) (*xml.StartElement, error) {

	// Write start of outgoing doc.
	buf := new(bytes.Buffer)
	if err := writeXMLStartElement(buf, start); err != nil {
		return nil, err
	}
	if err := stream.send(buf.Bytes()); err != nil {
		return nil, err
	}

	// Read and return start of incoming doc.
	rstart, err := nextStartElement(stream.dec)
	if err != nil {
		return nil, err
	}

	// Collect top-level namespaces.
	stream.incomingNamespace = make(nsMap)
	for _, attr := range rstart.Attr {
		if attr.Name.Space == "xmlns" {
			stream.incomingNamespace[attr.Value] = attr.Name.Local
		} else if attr.Name.Space == "" && attr.Name.Local == "xmlns" {
			stream.incomingNamespace[attr.Value] = ""
		}
	}

	return rstart, nil
}

// Send the end element that closes the stream.
func (stream *Stream) SendEnd(end *xml.EndElement) error {
	buf := new(bytes.Buffer)
	if err := writeXMLEndElement(buf, end); err != nil {
		return err
	}
	if err := stream.send(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

// Send a stanza. Used to write a complete, top-level element.
func (stream *Stream) Send(v interface{}) error {
	if stream.config.LogStanzas {
		bytes, err := xml.Marshal(v)
		if err != nil {
			return err
		}
		return stream.send(bytes)
	}
	enc := xml.NewEncoder(stream.conn)
	return enc.Encode(v)
}

func (stream *Stream) send(b []byte) error {
	if stream.config.LogStanzas {
		log.Println("send:", string(b))
	}
	if _, err := stream.conn.Write(b); err != nil {
		return err
	}
	return nil
}

// Find start of next stanza.
// Bad things are very likely to happen if a call to Next() is successful but
// you don't actually decode or skip the element.
func (stream *Stream) Next() (*xml.StartElement, error) {

	start, err := nextStartElement(stream.dec)
	if err != nil {
		return nil, err
	}

	if stream.config.LogStanzas {
		if xml, err := collectElement(stream.dec, start, stream.incomingNamespace); err != nil {
			return nil, err
		} else {
			stream.stanzaBuf = xml
		}
		log.Println("recv:", stream.stanzaBuf)
	}

	return start, nil
}

func nextStartElement(dec *xml.Decoder) (*xml.StartElement, error) {
	for {
		t, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}
		switch e := t.(type) {
		case xml.StartElement:
			return &e, nil
		case xml.EndElement:
			log.Printf("EOF due to %s\n", e.Name)
			return nil, io.EOF
		}
	}
	panic("Unreachable")
}

// Skip reads tokens until it reaches the end element of the most recent start
// element that has already been read.
func (stream *Stream) Skip() error {
	if stream.config.LogStanzas && stream.stanzaBuf != "" {
		stream.stanzaBuf = ""
		return nil
	}
	return stream.dec.Skip()
}

// Decode a stanza.
// If start is not nil, the stanza for the start element that's already been
// consumed is read. A nil start will read the next stanza in the stream.
// See xml.Decoder.DecodeElement for decoding rules.
func (stream *Stream) Decode(v interface{}, start *xml.StartElement) error {

	// Explicity lookup next start element to ensure stream is validated,
	// stanza is logged, etc.
	if start == nil {
		if se, err := stream.Next(); err != nil {
			return err
		} else {
			start = se
		}
	}

	if stream.config.LogStanzas {
		return xml.Unmarshal([]byte(stream.stanzaBuf), v)
	}

	return stream.dec.DecodeElement(v, start)
}

// Collect the element with the start that's already been consumed into a
// buffer. Namespaces are munged so the buffer can be correctly parsed outside
// the context of the stream. This is used for logging the received data.
func collectElement(dec *xml.Decoder, start *xml.StartElement, nsmap nsMap) (string, error) {

	var collector struct {
		XML []byte `xml:",innerxml"`
	}

	if err := dec.DecodeElement(&collector, start); err != nil {
		return "", err
	}

	name := start.Name
	attrs := start.Attr

	// Map the element's namespace.
	if ns, ok := nsmap[name.Space]; ok {
		// Element's namespace is one of the stream namespaces. Update the
		// element's namespace and add the namespace to the element's attrs.
		attrs = append(attrs, xml.Attr{xml.Name{"xmlns", ns}, name.Space})
		name = xml.Name{ns, name.Local}
	} else {
		// Check that Go's xml package hasn't duplicated the default ns as the
		// element name's space. If so, clear it.
		for _, attr := range attrs {
			if attr.Name == (xml.Name{"", "xmlns"}) {
				if name.Space == attr.Value {
					name = xml.Name{"", start.Name.Local}
				}
				break
			}
		}
	}

	start = &xml.StartElement{name, attrs}

	buf := new(bytes.Buffer)
	if err := writeXMLStartElement(buf, start); err != nil {
		return "", err
	}
	if _, err := buf.Write(collector.XML); err != nil {
		return "", err
	}
	if err := writeXMLEndElement(buf, &xml.EndElement{start.Name}); err != nil {
		return "", err
	}

	return buf.String(), nil
}
