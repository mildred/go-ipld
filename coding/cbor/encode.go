package cbor

import (
	"io"

	memory "github.com/ipfs/go-ipld/memory"
	cbor "github.com/whyrusleeping/cbor/go"
)

// Encode to CBOR, add the multicodec header
func Encode(w io.Writer, node memory.Node) error {
	_, err := w.Write(Header)
	if err != nil {
		return err
	}

	return RawEncode(w, node)
}

// Encode to CBOR, do not add the multicodec header
func RawEncode(w io.Writer, node memory.Node) error {
	return cbor.NewEncoder(w).Encode(node)
}
