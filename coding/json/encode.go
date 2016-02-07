package json

import (
	"encoding/json"
	"io"

	memory "github.com/ipfs/go-ipld/memory"
)

// Encode to JSON, add the multicodec header
func Encode(w io.Writer, node memory.Node) error {
	_, err := w.Write(Header)
	if err != nil {
		return err
	}

	return RawEncode(w, node)
}

// Encode to JSON, do not add the multicodec header
func RawEncode(w io.Writer, node memory.Node) error {
	return json.NewEncoder(w).Encode(node)
}
