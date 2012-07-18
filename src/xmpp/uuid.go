package xmpp

import (
	"crypto/rand"
	"fmt"
)

// Generate a UUID4.
func UUID4() string {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		panic(err)
	}
	uuid[6] = (uuid[6] & 0x0F) | 0x40
	uuid[8] = (uuid[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
