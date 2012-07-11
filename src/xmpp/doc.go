/*
	Package for implementing XMPP clients and components.

	The package is built around the concept of an XML stream - a pair of XML
	documents written to and read from a TCP connection. Top-level elements in
	the document form the messages processed by either end of the connection.

	An XML stream is then configured for an XMPP conversation, as either a
	client (chat, etc) or component (a sort of server plugin).

	Create a client:

		jid, err := xmpp.ParseJID("alice@wonderland.lit/some-resource")
		addr, err := xmpp.HomeServerAddrs(jid)
		stream, err := xmpp.NewStream(addr[0], nil)
		X, err := xmpp.NewClientXMPP(stream, jid, "password", nil)

	Create a component:

		jid, err := xmpp.ParseJID("rabbithole.wonderland.lit")
		stream, err := xmpp.NewStream("localhost:5347", nil)
		X, err := xmpp.NewComponentXMPP(stream, jid, "secret")

	Messages are sent using the XMPP.Send method, e.g. a client typically
	announces its presence on the XMPP network as soon as it's connected:

		X.Send(xmpp.Presence{})

	Incoming messages can be received in a simple loop, ended by an os.EOF for
	clean shutdown or any other error for something unexpected. XMPP defines
	four types of stanza: <error/>, <iq/>, <message/> and <presence/>
	represented by Error, Iq, Message and Presence structs respectively.

		for {
			stanza, err := X.Recv()
			if err == io.EOF {
				break
			}
			log.Printf("%T : %v\n", stanza, stanza)
		}

	Note: A "bound" JID is negotatiated during XMPP setup and may be different
	to the JID passed to the New(Client|Component)XMPP() call. Always
	use the XMPP instance's JID attribute in any stanzas.
*/
package xmpp
