package xmpp

import "encoding/base64"

func saslEncodePlain(user, password string) string {
	return base64.StdEncoding.EncodeToString([]byte("\x00" + user + "\x00" + password))
}
