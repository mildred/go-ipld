package ipld

import (
	"testing"
	"reflect"

	mh "github.com/jbenet/go-multihash"
)

type TC struct {
	src    Node
	json   Node
	jsonld Node
	links  map[string]string
	typ    string
	ctx    interface{}
}

var testCases []TC

func mmh(b58 string) mh.Multihash {
	h, err := mh.FromB58String(b58)
	if err != nil {
		panic("failed to decode multihash")
	}
	return h
}

func init() {
	testCases = append(testCases, TC{
		src: Node{
			"foo": "bar",
			"bar": []int{1, 2, 3},
			"baz": Node{
				"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
		},
		json: Node{
			"foo": "bar",
			"bar": []int{1, 2, 3},
			"baz": Node{
				"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
		},
		jsonld: Node{
			"foo": "bar",
			"bar": []int{1, 2, 3},
			"baz": Node{
				"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
		},
		links: map[string]string{
			"baz": "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
		},
		typ: "",
		ctx: nil,
	}, TC{
		src: Node{
			"foo": "bar",
			"bar": []int{1, 2, 3},
			"@container": "@index",
			"@index": "links",
			"baz": Node{
				"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
		},
		json: Node{
			"foo": "bar",
			"bar": []int{1, 2, 3},
			"baz": Node{
				"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
		},
		jsonld: Node{
			"links": Node{
				"foo": "bar",
				"bar": []int{1, 2, 3},
				"baz": Node{
					"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
				},
			},
		},
		links: map[string]string{
			"baz": "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
		},
		typ: "",
		ctx: nil,
	}, TC{
		src: Node{
			"@attrs": Node{
				"attr": "val",
			},
			"foo":        "bar",
			"@index":     "files",
			"@type":      "commit",
			"@container": "@index",
			"@context": "/ipfs/QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo/mdag",
			"baz": Node{
				"foobar": "barfoo",
				"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
			"\\@bazz": Node{
				"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
			"bar/ra\\b": Node{
				"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPb",
			},
			"bar": Node{
				"@container": "@index",
				"foo": Node{
					"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPa",
				},
			},
		},
		json: Node{
			"attr": "val",
			"files": Node{
				"foo": "bar",
				"baz": Node{
					"foobar": "barfoo",
					"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
				},
				"@bazz": Node{
					"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
				},
				"bar/ra\\b": Node{
					"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPb",
				},
				"bar": Node{
					"foo": Node{
						"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPa",
					},
				},
			},
		},
		jsonld: Node{
			"attr": "val",
			"@type": "commit",
			"@context": "/ipfs/QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo/mdag",
			"files": Node{
				"foo":        "bar",
				"baz": Node{
					"foobar": "barfoo",
					"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
				},
				"@bazz": Node{
					"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
				},
				"bar/ra\\b": Node{
					"mlink":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPb",
				},
				"bar": Node{
				},
			},
		},
		links: map[string]string{
			"files/baz":           "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			"files/@bazz":         "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			"files/bar\\/ra\\\\b": "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPb",
			"files/bar/foo":       "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPa",
		},
		typ: "",
		ctx: "/ipfs/QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo/mdag",
	})
}

func TestParsing(t *testing.T) {
	for tci, tc := range testCases {
		t.Logf("===== Test case #%d =====", tci)
		doc := tc.src

		// check links
		links := doc.Links()
		t.Logf("links: %#v", links)
		if len(links) != len(tc.links) {
			t.Errorf("links do not match, not the same number of links, expected %d, got %d", len(tc.links), len(links))
		}
		for k, l1 := range tc.links {
			l2 := links[k]
			if l1 != l2["mlink"] {
				t.Errorf("links do not match. %d/%#v %#v != %#v[mlink]", tci, k, l1, l2)
			}
		}

		// check JSON mode
		json := doc.StripDirectivesAll()
		if !reflect.DeepEqual(tc.json, json) {
			t.Errorf("JSON version mismatch.\nGot:    %#v\nExpect: %#v", json, tc.json)
		} else {
			t.Log("JSON version OK")
		}

		// check JSON-LD mode
		jsonld := doc.ToLinkedDataAll()
		if !reflect.DeepEqual(tc.jsonld, jsonld) {
			t.Errorf("JSON-LD version mismatch.\nGot:    %#v\nExpect: %#v", jsonld, tc.jsonld)
		} else {
			t.Log("JSON-LD version OK")
		}

	}
}
