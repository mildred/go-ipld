package ipld

import (
	"errors"
	"reflect"
	"sort"

	"github.com/ipfs/go-ipld/reader"
	mh "github.com/jbenet/go-multihash"
)

// These are the constants used in the format.
const (
	IDKey    = "@id"      // the id of the object (JSON-LD)
	TypeKey  = "@type"    // the type of the object (JSON-LD)
	ValueKey = "@value"   // the value of the object (JSON-LD)
	CtxKey   = "@context" // the JSON-LD style context

	CodecKey = "@codec" // used to determine which multicodec to use
	LinkKey  = "mlink"  // key for merkle-links
)

// Node is an IPLD node. effectively, it is equivalent to a JSON-LD object.
// (which is {,de}serialized to CBOR or JSON) which derives from a base
// schema, the IPLD schema (@context). This allows keys to specify:
//
//    "myfield": { "@value": "Qmabcbcbdba", "@type": "mlink" }
//
// "mlink" signals that "@value" is taken to be a merkle-link, which IPFS
// handles specially.
type Node map[string]interface{}

// Get retrieves a property of the node. it uses unix path notation,
// splitting on "/".
func (n Node) Get(path_ string) interface{} {
	return GetPath(n, path_)
}

// Type is a convenience method to retrieve "@type", if there is one.
func (d Node) Type() string {
	s, _ := d[TypeKey].(string)
	return s
}

// Context is a convenience method to retrieve the JSON-LD-style context.
// It may be a string (link to context), a []interface (multiple contexts),
// or a Node (an inline context)
func (d Node) Context() interface{} {
	return d[CtxKey]
}

// Links returns all the merkle-links in the document. When the document
// is parsed, all the links are identified and references are cached, so
// getting the links only walks the document _once_. Note though that the
// entire document must be walked.
func (d Node) Links() map[string]Link {
	return Links(d)
}

// Link is a merkle-link to a target Node. The Link object is
// represented by a JSON-LD style map:
//
//   { "@type": "mlink", "@value": <multihash>, ... }
//
// Links support adding other data, which will be
// serialized and de-serialized along with the link.
// This allows users to set other properties on links:
//
//   {
//     "@type": "mlink",
//     "@value": <multihash>,
//     "unixType": "dir",
//     "unixMode": "0777",
//   }
//
// looking at a whole filesystem node, we might see something like:
//
//   {
//     "@context": "/ipfs/Qmf1ec6n9f8kW8JTLjqaZceJVpDpZD4L3aPoJFvssBE7Eb/merkleweb",
//     "foo": {
//       "@type": "mlink",
//       "@value": <multihash>,
//       "unixType": "dir",
//       "unixMode": "0777",
//     },
//     "bar": {
//       "@type": "mlink",
//       "@value": <multihash>,
//       "unixType": "file",
//       "unixMode": "0755",
//     }
//   }
//
type Link Node

// Type returns the type of the link. It should be "mlink"
func (l Link) Type() string {
	s, _ := l[TypeKey].(string)
	return s
}

// HashStr returns the string value of l["mlink"],
// which is the value we use to store hashes.
func (l Link) LinkStr() string {
	s, _ := l[LinkKey].(string)
	return s
}

// Hash returns the multihash value of the link.
func (l Link) Hash() (mh.Multihash, error) {
	s := l.LinkStr()
	if s == "" {
		return nil, errors.New("no hash in link")
	}
	return mh.FromB58String(s)
}

// Equal returns whether two Link objects are equal.
// It uses reflect.DeepEqual, so beware compating
// large structures.
func (l Link) Equal(l2 Link) bool {
	return reflect.DeepEqual(l, l2)
}

// Links walks given node and returns all links found,
// in a flattened map. the map keys use path notation,
// made up of the intervening keys. For example:
//
// 		{
//			"foo": {
//				"quux": { @type: mlink, @value: Qmaaaa... },
// 			},
//			"bar": {
//				"baz": { @type: mlink, @value: Qmbbbb... },
//			},
//		}
//
// would produce links:
//
// 		{
//			"foo/quux": { @type: mlink, @value: Qmaaaa... },
//			"bar/baz": { @type: mlink, @value: Qmbbbb... },
//		}
//
// WARNING: your nodes should not use `/` as key names. it will
// confuse link parsers. thus, if we find any map keys with slash
// in them, we simply ignore them.
func Links(n Node) map[string]Link {
	m := map[string]Link{}
	Walk(n, func(root, curr Node, path string, err error) error {
		if err != nil {
			return err // if anything went wrong, bail.
		}

		if l, ok := LinkCast(curr); ok {
			m[path] = l
		}
		return nil
	})
	return m
}

// checks whether a value is a link. for now we assume that all links
// follow:
//
//   { "mlink": "<multihash>" }
func IsLink(v interface{}) bool {
	vn, ok := v.(Node)
	if !ok {
		return false
	}

	_, ok = vn[LinkKey].(string)
	return ok
}

// returns the link value of an object. for now we assume that all links
// follow:
//
//   { "mlink": "<multihash>" }
func LinkCast(v interface{}) (l Link, ok bool) {
	if !IsLink(v) {
		return
	}

	l = make(Link)
	for k, v := range v.(Node) {
		l[k] = v
	}
	return l, true
}

func (n Node) Read(fun reader.ReadFun) error {
	err := read(n, fun, []interface{}{})
	if err == reader.NodeReadAbort {
		err = nil
	}
	return err
}

func read(curr interface{}, fun reader.ReadFun, path []interface{}) error {
	if nc, ok := curr.(Node); ok { // it's a node!
		err := fun(path, reader.TokenNode, nil)
		if err == reader.NodeReadSkip {
			return nil
		} else if err != nil {
			return err
		}

		// Iterate in fixed order (by default, go randomize iteration order)
		// Simulate reading from a file where the order is fixed
		keys := make([]string, 0, len(nc))
		for k := range nc {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			err := fun(path, reader.TokenKey, k)
			if err == reader.NodeReadSkip {
				return nil
			} else if err != nil {
				return err
			}

			subpath := append(path, k)
			err = read(nc[k], fun, subpath)
			if err != nil {
				return err
			}
		}

		err = fun(path, reader.TokenEndNode, nil)
		if err == reader.NodeReadSkip {
			return nil
		} else if err != nil {
			return err
		}

	} else if sc, ok := curr.([]interface{}); ok { // it's a slice!
		err := fun(path, reader.TokenArray, nil)
		if err == reader.NodeReadSkip {
			return nil
		} else if err != nil {
			return err
		}

		for i, v := range sc {
			err := fun(path, reader.TokenIndex, i)
			if err == reader.NodeReadSkip {
				return nil
			} else if err != nil {
				return err
			}

			subpath := append(path, i)
			err = read(v, fun, subpath)
			if err != nil {
				return err
			}
		}

		err = fun(path, reader.TokenEndArray, nil)
		if err == reader.NodeReadSkip {
			return nil
		} else if err != nil {
			return err
		}

	} else {
		err := fun(path, reader.TokenValue, curr)
		if err == reader.NodeReadSkip {
			return nil
		} else if err != nil {
			return err
		}
	}
	return nil
}
