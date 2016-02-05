package ipfsld

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	ipld "github.com/ipfs/go-ipld"
	reader "github.com/ipfs/go-ipld/reader"
	readertest "github.com/ipfs/go-ipld/reader/test"

	mc "github.com/jbenet/go-multicodec"
	mctest "github.com/jbenet/go-multicodec/test"
	assrt "github.com/mildred/assrt"
)

var codedFiles map[string][]byte = map[string][]byte{
	"json.testfile": []byte{},
	"cbor.testfile": []byte{},
}

func init() {
	for fname := range codedFiles {
		var err error
		codedFiles[fname], err = ioutil.ReadFile(fname)
		if err != nil {
			panic("could not read " + fname + ". please run: make " + fname)
		}
	}
}

type TC struct {
	cbor  []byte
	src   ipld.Node
	links map[string]ipld.Link
	typ   string
	ctx   interface{}
}

var testCases []TC

func init() {
	testCases = append(testCases, TC{
		[]byte{},
		ipld.Node{
			"foo": "bar",
			"bar": []int{1, 2, 3},
			"baz": ipld.Node{
				"@type": "mlink",
				"hash":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
		},
		map[string]ipld.Link{
			"baz": {"@type": "mlink", "hash": ("QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo")},
		},
		"",
		nil,
	})

	testCases = append(testCases, TC{
		[]byte{},
		ipld.Node{
			"foo":      "bar",
			"@type":    "commit",
			"@context": "/ipfs/QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo/mdag",
			"baz": ipld.Node{
				"@type": "mlink",
				"hash":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
			"bazz": ipld.Node{
				"@type": "mlink",
				"hash":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
			"bar": ipld.Node{
				"@type": "mlinkoo",
				"hash":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
			"bar2": ipld.Node{
				"foo": ipld.Node{
					"@type": "mlink",
					"hash":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
				},
			},
		},
		map[string]ipld.Link{
			"baz":      {"@type": "mlink", "hash": ("QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo")},
			"bazz":     {"@type": "mlink", "hash": ("QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo")},
			"bar2/foo": {"@type": "mlink", "hash": ("QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo")},
		},
		"",
		"/ipfs/QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo/mdag",
	})

}

func TestHeaderMC(t *testing.T) {
	codec := Multicodec()
	for _, tc := range testCases {
		mctest.HeaderTest(t, codec, &tc.src)
	}
}

func TestRoundtripBasicMC(t *testing.T) {
	codec := Multicodec()
	for _, tca := range testCases {
		var tcb ipld.Node
		mctest.RoundTripTest(t, codec, &(tca.src), &tcb)
	}
}

// Test decoding and encoding a json and cbor file
func TestCodecsDecodeEncode(t *testing.T) {
	for fname, testfile := range codedFiles {
		var n ipld.Node
		codec := Multicodec()

		if err := mc.Unmarshal(codec, testfile, &n); err != nil {
			t.Log(testfile)
			t.Error(err)
			continue
		}

		linksExpected := map[string]ipld.Link{
			"abc": ipld.Link{
				"mlink": "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V",
			},
		}
		linksActual := ipld.Links(n)
		if !reflect.DeepEqual(linksExpected, linksActual) {
			t.Logf("Expected: %#v", linksExpected)
			t.Logf("Actual:   %#v", linksActual)
			t.Logf("node: %#v\n", n)
			t.Error("Links are not expected in " + fname)
			continue
		}

		encoded, err := mc.Marshal(codec, &n)
		if err != nil {
			t.Error(err)
			return
		}

		if !bytes.Equal(testfile, encoded) {
			t.Error("marshalled values not equal in " + fname)
			t.Log(string(testfile))
			t.Log(string(encoded))
			t.Log(testfile)
			t.Log(encoded)
		}
	}
}

func TestStream(t *testing.T) {
	a := assrt.NewAssert(t)
	json, err := Decode(bytes.NewReader(codedFiles["json.testfile"]))
	a.MustNil(err)

	t.Logf("Reading json.testfile")
	readertest.CheckReader(t, json, []readertest.Callback{
		readertest.Callback{[]interface{}{}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{}, reader.TokenKey, "@codec"},
		readertest.Callback{[]interface{}{"@codec"}, reader.TokenValue, "/json"},
		readertest.Callback{[]interface{}{}, reader.TokenKey, "abc"},
		readertest.Callback{[]interface{}{"abc"}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{"abc"}, reader.TokenKey, "mlink"},
		readertest.Callback{[]interface{}{"abc", "mlink"}, reader.TokenValue, "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V"},
		readertest.Callback{[]interface{}{"abc"}, reader.TokenEndNode, nil},
		readertest.Callback{[]interface{}{}, reader.TokenEndNode, nil},
	})

	cbor, err := Decode(bytes.NewReader(codedFiles["cbor.testfile"]))
	a.MustNil(err)

	t.Logf("Reading cbor.testfile")
	readertest.CheckReader(t, cbor, []readertest.Callback{
		readertest.Callback{[]interface{}{}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{}, reader.TokenKey, "abc"},
		readertest.Callback{[]interface{}{"abc"}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{"abc"}, reader.TokenKey, "mlink"},
		readertest.Callback{[]interface{}{"abc", "mlink"}, reader.TokenValue, "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V"},
		readertest.Callback{[]interface{}{"abc"}, reader.TokenEndNode, nil},
		readertest.Callback{[]interface{}{}, reader.TokenEndNode, nil},
	})
}
