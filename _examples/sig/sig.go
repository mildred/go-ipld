package sig

import (
	"github.com/ipfs/go-ipld/links"
)

// Signature is an object that represents a cryptographic
// signature on any other merkle node.
//
// this would serialize to:
//
//   {
//    "@context": "/ipfs/<hash-of-schema>/signature"
//   	"key": { "@link": "<hash1>" },
//   	"object": { "@link": "<hash2>" },
//   	"sig": "<sign(sk, <hash2>)>"
//   }
//
type Signature struct {
	Key    links.SimpleHashLink // the signing key
	Object links.SimpleHashLink // what is signed
	Sig    []byte               // the data representing the signature
}

// Sign creates a signature from a given key and a link to data.
// Since this is a merkledag, signing the link is effectively the
// same as an hmac signature.
func Sign(skey key.SigningKey, signed mh.Multihash) (*Signature, error) {
	sig, err := skey.Sign(signed)
	if err != nil {
		return nil, err
	}

	return &Signature{
		Key:    links.SimpleHashLink{Hash: key.Hash},
		Object: links.SimpleHashLink{Hash: signed},
		Sig:    sig,
	}, nil
}
