package xmpp

import (
	"crypto/sha1"
	"encoding/xml"
	"errors"
	"fmt"
)

// Create a component XMPP connection.
func NewComponentXMPP(addr string, jid JID, secret string) (*XMPP, error) {

	stream, err := NewStream(addr)
	if err != nil {
		return nil, err
	}

	streamId, err := startComponent(stream, jid)
	if err != nil {
		return nil, err
	}

	if err := handshake(stream, streamId, secret); err != nil {
		return nil, err
	}

	return newXMPP(jid, stream), nil
}

func startComponent(stream *Stream, jid JID) (string, error) {

	start := xml.StartElement{
		xml.Name{"stream", "stream"},
		[]xml.Attr{
			xml.Attr{xml.Name{"", "xmlns"}, "jabber:component:accept"},
			xml.Attr{xml.Name{"xmlns", "stream"}, "http://etherx.jabber.org/streams"},
			xml.Attr{xml.Name{"", "to"}, jid.Full()},
		},
	}

	if err := stream.SendStart(&start); err != nil {
		return "", err
	}

	streamId := ""
	if e, err := stream.Next(&xml.Name{nsStream, "stream"}); err != nil {
		return "", err
	} else {
		// Find the stream id.
		for _, attr := range e.Attr {
			if attr.Name.Local == "id" {
				streamId = attr.Value
				break
			}
		}
		if streamId == "" {
			return "", errors.New("Missing stream id")
		}
	}

	return streamId, nil
}

func handshake(stream *Stream, streamId, secret string) error {

	hash := sha1.New()
	hash.Write([]byte(streamId))
	hash.Write([]byte(secret))

	// Send handshake.
	handshake := saslHandshake{Value: fmt.Sprintf("%x", hash.Sum(nil))}
	if err := stream.Send(&handshake); err != nil {
		return err
	}

	// Get handshake response.
	if _, err := stream.Next(&xml.Name{"jabber:component:accept", "handshake"}); err != nil {
		return err
	}

	return nil
}

type saslHandshake struct {
	XMLName xml.Name `xml:"handshake"`
	Value string `xml:",chardata"`
}
