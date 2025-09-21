package data

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
)

type KeyLength struct {
	EncodedLen int
	DecodedLen int
}

func GetDecodedLen(k int) int { return base64.StdEncoding.DecodedLen(int(k)) }

var key64 = KeyLength{64, GetDecodedLen(64)}

type Secret []byte

func (s Secret) PopulateRand() {
	rand.Read(s) //nolint:all rand.Read never returns error except on legacy linux systems
}

func (s Secret) PopulateFromBase64(src []byte) {
	base64.StdEncoding.Decode(s, src) //nolint:all Decode won't return an error since the size is controlled
}

func (s Secret) GetHash() [64]byte {
	return sha512.Sum512(s)
}

func (s Secret) GetBase64encoded(out []byte) {
	base64.StdEncoding.Encode(out, s)
}
