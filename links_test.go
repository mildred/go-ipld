package ipld

import (
	"testing"

	links "github.com/ipfs/go-ipld/links"
	memory "github.com/ipfs/go-ipld/memory"
	mh "github.com/jbenet/go-multihash"
	assrt "github.com/mildred/assrt"
)

const (
	h1 string = "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPa"
	h2 string = "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPb"
	h3 string = "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPc"
)

var (
	mh1 mh.Multihash = FromB58String(h1)
	mh2 mh.Multihash = FromB58String(h2)
	mh3 mh.Multihash = FromB58String(h3)
)

func FromB58String(s string) (h mh.Multihash) {
	h, err := mh.FromB58String(s)
	if err != nil {
		panic(err)
	}
	return h
}

type Link1 struct {
	Hash []byte `ipld:"multihash"` //mh.Multihash
	Name string `ipld:"name"`
	Size uint64 `ipld:"key:size"`
}

func TestParsing(t *testing.T) {
	a := assrt.NewAssert(t)

	n := memory.Node{
		"foo": memory.Node{
			links.LinkKey: h1,
			"size":        3,
		},
		"bar": memory.Node{
			"baz": memory.Node{
				links.LinkKey: h2,
				"size":        42,
				"boo": memory.Node{
					links.LinkKey: h3,
				},
			},
		},
	}

	var links []Link1

	err := ReadLinks(n, &links)
	a.Nil(err)

	a.Equal([]Link1{
		Link1{
			Name: "boo",
			Hash: mh3,
			Size: 0,
		},
		Link1{
			Name: "baz",
			Hash: mh2,
			Size: 42,
		},
		Link1{
			Name: "foo",
			Hash: mh1,
			Size: 3,
		},
	}, links)
}

type Link2 struct {
	links.BaseLink
	Size uint64 `ipld:"key:size"`
}

func TestParsing2(t *testing.T) {
	a := assrt.NewAssert(t)

	n := memory.Node{
		"foo": memory.Node{
			links.LinkKey: h1,
			"size":        3,
		},
		"bar": memory.Node{
			"baz": memory.Node{
				links.LinkKey: h2,
				"size":        42,
				"boo": memory.Node{
					links.LinkKey: h3,
				},
			},
		},
	}

	var convlinks []Link2

	err := ReadLinks(n, &convlinks)
	a.Nil(err)

	a.Equal([]Link2{
		Link2{
			BaseLink: links.BaseLink{
				Name:       "boo",
				Hash:       mh3,
				Link:       h3,
				StringPath: []string{"bar", "baz", "boo"},
				Path:       []interface{}{"bar", "baz", "boo"},
			},
			Size: 0,
		},
		Link2{
			BaseLink: links.BaseLink{
				Name:       "baz",
				Hash:       mh2,
				Link:       h2,
				StringPath: []string{"bar", "baz"},
				Path:       []interface{}{"bar", "baz"},
			},
			Size: 42,
		},
		Link2{
			BaseLink: links.BaseLink{
				Name:       "foo",
				Hash:       mh1,
				Link:       h1,
				StringPath: []string{"foo"},
				Path:       []interface{}{"foo"},
			},
			Size: 3,
		},
	}, convlinks)
}
