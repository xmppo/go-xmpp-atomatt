package xmpp

import (
	"encoding/xml"
	"fmt"
	"io"
)

type nsMap map[string]string

// Write an xml.StartElement.
func writeXMLStartElement(w io.Writer, start *xml.StartElement) error {
	if _, err := w.Write([]byte{'<'}); err != nil {
		return err
	}
	if err := writeXMLName(w, start.Name); err != nil {
		return err
	}
	for _, attr := range start.Attr {
		if _, err := w.Write([]byte{' '}); err != nil {
			return err
		}
		if err := writeXMLAttr(w, attr); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte{'>'}); err != nil {
		return err
	}
	return nil
}

// Write an xml.StartElement.
func writeXMLEndElement(w io.Writer, end *xml.EndElement) error {
	if _, err := w.Write([]byte{'<', '/'}); err != nil {
		return err
	}
	if err := writeXMLName(w, end.Name); err != nil {
		return err
	}
	if _, err := w.Write([]byte{'>'}); err != nil {
		return err
	}
	return nil
}

// Write a xml.Name.
func writeXMLName(w io.Writer, name xml.Name) error {
	if name.Space == "" {
		if _, err := fmt.Fprintf(w, name.Local); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "%s:%s", name.Space, name.Local); err != nil {
			return err
		}
	}
	return nil
}

// Write a xml.Attr.
func writeXMLAttr(w io.Writer, attr xml.Attr) error {
	if err := writeXMLName(w, attr.Name); err != nil {
		return err
	}
	if _, err := w.Write([]byte{'=', '\''}); err != nil {
		return err
	}
	xml.Escape(w, []byte(attr.Value))
	if _, err := w.Write([]byte{'\''}); err != nil {
		return err
	}
	return nil
}

func startElementIter(dec *xml.Decoder) func() *xml.StartElement {
	return func() *xml.StartElement {
		for {
			if tok, err := dec.Token(); err != nil {
				return nil
			} else {
				if start, ok := tok.(xml.StartElement); ok {
					return &start
				}
			}
		}
		return nil
	}
}
