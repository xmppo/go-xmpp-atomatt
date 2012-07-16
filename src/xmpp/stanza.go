package xmpp

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

// XMPP <iq/> stanza.
type Iq struct {
	XMLName xml.Name `xml:"iq"`
	Id string `xml:"id,attr"`
	Type string `xml:"type,attr"`
	To string `xml:"to,attr,omitempty"`
	From string `xml:"from,attr,omitempty"`
	Payload string `xml:",innerxml"`
	Error *Error `xml:"error"`
}

// Encode the value to an XML string and set as the payload. See xml.Marshal
// for how the value is encoded.
func (iq *Iq) PayloadEncode(v interface{}) error {
	bytes, err := xml.Marshal(v)
	if err != nil {
		return err
	}
	iq.Payload = string(bytes)
	return nil
}

// Decode the payload (an XML string) into the given value. See xml.Unmarshal
// for how the value is decoded.
func (iq *Iq) PayloadDecode(v interface{}) error {
	return xml.Unmarshal([]byte(iq.Payload), v)
}

// Return the name of the payload element.
func (iq *Iq) PayloadName() (name xml.Name) {
	dec := xml.NewDecoder(bytes.NewBufferString(iq.Payload))
	tok, err := dec.Token()
	if err != nil {
		return
	}
	start, ok := tok.(xml.StartElement)
	if !ok {
		return
	}
	return start.Name
}

// XMPP <message/> stanza.
type Message struct {
	XMLName xml.Name `xml:"message"`
	Id string `xml:"id,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	To string `xml:"to,attr,omitempty"`
	From string `xml:"from,attr,omitempty"`
	Subject string `xml:"subject,omitempty"`
	Body string `xml:"body,omitempty"`
}

// XMPP <presence/> stanza.
type Presence struct {
	XMLName xml.Name `xml:"presence"`
	Id string `xml:"id,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	To string `xml:"to,attr,omitempty"`
	From string `xml:"from,attr,omitempty"`
}

// XMPP <error/>. May occur as a top-level stanza or embedded in another
// stanza, e.g. an <iq type="error"/>.
type Error struct {
	XMLName xml.Name `xml:"error"`
	Type string `xml:"type,attr"`
	Text string `xml:"text"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Text)
}
