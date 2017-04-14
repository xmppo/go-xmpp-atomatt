package xmpp

import (
	"crypto/sha1"
	"encoding/xml"
	"errors"
	"fmt"
)

// Create a component XMPP connection over the stream.
func NewComponentXMPP(stream *Stream, jid JID, secret string) (*XMPP, error) {

	streamID, err := startComponent(stream, jid)
	if err != nil {
		return nil, err
	}

	if err := handshake(stream, streamID, secret); err != nil {
		return nil, err
	}

	return newXMPP(jid, stream), nil
}

func startComponent(stream *Stream, jid JID) (string, error) {

	start := xml.StartElement{
		xml.Name{"stream", "stream"},
		[]xml.Attr{
			xml.Attr{xml.Name{"", "xmlns"}, nsComponentAccept},
			xml.Attr{xml.Name{"xmlns", "stream"}, nsStreams},
			xml.Attr{xml.Name{"", "to"}, jid.Full()},
		},
	}

	var streamID string

	rstart, err := stream.SendStart(&start)
	if err != nil {
		return "", err
	}
	if rstart.Name != (xml.Name{nsStreams, "stream"}) {
		return "", fmt.Errorf("unexpected start element: %s", rstart.Name)
	}
	// Find the stream id.
	for _, attr := range rstart.Attr {
		if attr.Name.Local == "id" {
			streamID = attr.Value
			break
		}
	}
	if streamID == "" {
		return "", errors.New("Missing stream id")
	}

	return streamID, nil
}

func handshake(stream *Stream, streamID, secret string) error {

	hash := sha1.New()
	hash.Write([]byte(streamID))
	hash.Write([]byte(secret))

	// Send handshake.
	handshake := saslHandshake{Value: fmt.Sprintf("%x", hash.Sum(nil))}
	if err := stream.Send(&handshake); err != nil {
		return err
	}

	// Get handshake response.
	start, err := stream.Next()
	if err != nil {
		return err
	}
	if start.Name != (xml.Name{nsComponentAccept, "handshake"}) {
		return fmt.Errorf("Expected <handshake/>, for %s", start.Name)
	}
	return stream.Skip()
}

type saslHandshake struct {
	XMLName xml.Name `xml:"jabber:component:accept handshake"`
	Value   string   `xml:",chardata"`
}
