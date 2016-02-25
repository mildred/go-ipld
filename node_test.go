package ipld

import (
	"testing"

	links "github.com/ipfs/go-ipld/links"
	memory "github.com/ipfs/go-ipld/memory"
	assrt "github.com/mildred/assrt"
)

type Node struct {
	Links []links.SimpleHashLink `ipld:"key:links"`
	Data  string                 `ipld:"key:data"`
}

func TestParsingNode(t *testing.T) {
	a := assrt.NewAssert(t)

	n := memory.Node{
		"links": []interface{}{
			memory.Node{
				links.LinkKey: h1,
			},
			memory.Node{
				links.LinkKey: h2,
			},
			memory.Node{
				links.LinkKey: h3,
			},
		},
		"data": "foobar",
	}

	var node Node

	err := Unmarshal(n, &node)
	a.Nil(err)

	t.Logf("%#v", node)
	a.Equal(Node{
		Links: []links.SimpleHashLink{
			links.SimpleHashLink{Hash: mh1},
			links.SimpleHashLink{Hash: mh2},
			links.SimpleHashLink{Hash: mh3},
		},
		Data: "foobar",
	}, node)
}
