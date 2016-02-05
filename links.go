package ipld

import (
	"fmt"
	mh "github.com/ipfs/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-multihash"
)

const (
	LinkKey = "@link"
)

// Read the hash of an IPLD link.
func ReadLinkPath(value interface{}) (mh.Multihash, error) {
	svalue, has_value := value.(string)
	var h mh.Multihash
	if has_value {
		return mh.FromB58String(svalue)
	}
	return h, fmt.Errorf("Could not get multihash for %#v", value)
}
