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
		ldtest.Callback{[]interface{}{}, reader.TokenNode, nil},
		ldtest.Callback{[]interface{}{}, reader.TokenKey, "count"},
		ldtest.Callback{[]interface{}{"count"}, reader.TokenValue, 3},
		ldtest.Callback{[]interface{}{}, reader.TokenKey, "items"},
		ldtest.Callback{[]interface{}{"items"}, reader.TokenArray, nil},
		ldtest.Callback{[]interface{}{"items"}, reader.TokenIndex, 0},
		ldtest.Callback{[]interface{}{"items", 0}, reader.TokenValue, "a"},
		ldtest.Callback{[]interface{}{"items"}, reader.TokenIndex, 1},
		ldtest.Callback{[]interface{}{"items", 1}, reader.TokenValue, "b"},
		ldtest.Callback{[]interface{}{"items"}, reader.TokenIndex, 2},
		ldtest.Callback{[]interface{}{"items", 2}, reader.TokenValue, "c"},
		ldtest.Callback{[]interface{}{"items"}, reader.TokenEndArray, nil},
		ldtest.Callback{[]interface{}{}, reader.TokenKey, "key"},
		ldtest.Callback{[]interface{}{"key"}, reader.TokenValue, "value"},
		ldtest.Callback{[]interface{}{}, reader.TokenEndNode, nil},
	}

	ldtest.CheckReader(t, node, callbacks)
}
