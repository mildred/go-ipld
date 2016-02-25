package links

import (
	"testing"

	memory "github.com/ipfs/go-ipld/memory"
	assrt "github.com/mildred/assrt"
)

type Node struct {
	Links []SimpleHashLink `ipld:"key:links"`
	Data  string           `ipld:"key:data"`
}

func TestParsingNode(t *testing.T) {
	a := assrt.NewAssert(t)

	n := memory.Node{
		"links": []interface{}{
			memory.Node{
				LinkKey: h1,
			},
			memory.Node{
				LinkKey: h2,
			},
			memory.Node{
				LinkKey: h3,
			},
		},
		"data": "foobar",
	}

	var node Node

	err := Unmarshal(n, &node)
	a.Nil(err)

	a.Equal(Node{
		Links: []SimpleHashLink{
			SimpleHashLink{Hash: mh1},
			SimpleHashLink{Hash: mh2},
			SimpleHashLink{Hash: mh3},
		},
		Data: "foobar",
	}, node)
}
