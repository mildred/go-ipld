package cbor

import (
	"io"

	links "github.com/ipfs/go-ipld/links"
	memory "github.com/ipfs/go-ipld/memory"
	ma "github.com/jbenet/go-multiaddr"
	cbor "github.com/whyrusleeping/cbor/go"
)

// Encode to CBOR, add the multicodec header
func Encode(w io.Writer, node memory.Node, tags bool) error {
	_, err := w.Write(Header)
	if err != nil {
		return err
	}

	return RawEncode(w, node, tags)
}

// Encode to CBOR, do not add the multicodec header
func RawEncode(w io.Writer, node memory.Node, tags bool) error {
	enc := cbor.NewEncoder(w)
	if tags {
		enc.SetFilter(func(val interface{}) interface{} {
			var ok bool
			var node memory.Node
			var link interface{}
			var linkStr string
			var newNode memory.Node = memory.Node{}
			var linkPath interface{}
			var linkObject interface{}

			node, ok = val.(memory.Node)
			if !ok {
				return val
			}

			link, ok = node[links.LinkKey]
			if !ok {
				return val
			}

			linkStr, ok = link.(string)
			if !ok {
				return val
			}

			maddr, err := ma.NewMultiaddr(linkStr)
			for k, v := range node {
				if k != links.LinkKey {
					newNode[k] = v
				}
			}
			if err != nil || maddr.String() != linkStr {
				linkPath = linkStr
			} else {
				linkPath = maddr.Bytes()
			}

			if len(newNode) > 0 {
				linkObject = []interface{}{linkPath, newNode}
			} else {
				linkObject = linkPath
			}

			return &cbor.CBORTag{
				Tag:           TagIPLDLink,
				WrappedObject: linkObject,
			}
		})
	}
	return enc.Encode(node)
}
