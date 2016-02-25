package memory

import (
	"testing"

	reader "github.com/ipfs/go-ipld/stream"
	ldtest "github.com/ipfs/go-ipld/stream/test"
)

func TestReader(t *testing.T) {
	var node *Node

	node = &Node{
		"key":   "value",
		"items": []interface{}{"a", "b", "c"},
		"count": 3,
	}

	callbacks := []ldtest.Callback{
		ldtest.Cb(ldtest.Path(), reader.TokenNode, nil),
		ldtest.Cb(ldtest.Path(), reader.TokenKey, "count"),
		ldtest.Cb(ldtest.Path("count"), reader.TokenValue, 3),
		ldtest.Cb(ldtest.Path(), reader.TokenKey, "items"),
		ldtest.Cb(ldtest.Path("items"), reader.TokenArray, nil),
		ldtest.Cb(ldtest.Path("items"), reader.TokenIndex, 0),
		ldtest.Cb(ldtest.Path("items", 0), reader.TokenValue, "a"),
		ldtest.Cb(ldtest.Path("items"), reader.TokenIndex, 1),
		ldtest.Cb(ldtest.Path("items", 1), reader.TokenValue, "b"),
		ldtest.Cb(ldtest.Path("items"), reader.TokenIndex, 2),
		ldtest.Cb(ldtest.Path("items", 2), reader.TokenValue, "c"),
		ldtest.Cb(ldtest.Path("items"), reader.TokenEndArray, nil),
		ldtest.Cb(ldtest.Path(), reader.TokenKey, "key"),
		ldtest.Cb(ldtest.Path("key"), reader.TokenValue, "value"),
		ldtest.Cb(ldtest.Path(), reader.TokenEndNode, nil),
	}

	ldtest.CheckReader(t, node, callbacks)
}
