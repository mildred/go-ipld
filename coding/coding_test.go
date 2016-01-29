package coding

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	reader "github.com/ipfs/go-ipld"
	memory "github.com/ipfs/go-ipld/memory"
	readertest "github.com/ipfs/go-ipld/test"

	mc "github.com/jbenet/go-multicodec"
	mctest "github.com/jbenet/go-multicodec/test"
	assrt "github.com/mildred/assrt"
)

var codedFiles map[string][]byte = map[string][]byte{
	"json.testfile":     []byte{},
	"cbor.testfile":     []byte{},
	"protobuf.testfile": []byte{},
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
	src   memory.Node
	links map[string]memory.Link
	typ   string
	ctx   interface{}
}

var testCases []TC

func init() {
	testCases = append(testCases, TC{
		[]byte{},
		memory.Node{
			"foo": "bar",
			"bar": []int{1, 2, 3},
			"baz": memory.Node{
				"@type": "mlink",
				"hash":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
		},
		map[string]memory.Link{
			"baz": {"@type": "mlink", "hash": ("QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo")},
		},
		"",
		nil,
	})

	testCases = append(testCases, TC{
		[]byte{},
		memory.Node{
			"foo":      "bar",
			"@type":    "commit",
			"@context": "/ipfs/QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo/mdag",
			"baz": memory.Node{
				"@type": "mlink",
				"hash":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
			"bazz": memory.Node{
				"@type": "mlink",
				"hash":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
			"bar": memory.Node{
				"@type": "mlinkoo",
				"hash":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
			},
			"bar2": memory.Node{
				"foo": memory.Node{
					"@type": "mlink",
					"hash":  "QmZku7P7KeeHAnwMr6c4HveYfMzmtVinNXzibkiNbfDbPo",
				},
			},
		},
		map[string]memory.Link{
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
		var tcb memory.Node
		mctest.RoundTripTest(t, codec, &(tca.src), &tcb)
	}
}

// Test decoding and encoding a json and cbor file
func TestCodecsDecodeEncode(t *testing.T) {
	for _, fname := range []string{"json.testfile", "cbor.testfile"} {
		testfile := codedFiles[fname]
		var n memory.Node
		codec := Multicodec()

		if err := mc.Unmarshal(codec, testfile, &n); err != nil {
			t.Log(fname)
			t.Log(testfile)
			t.Error(err)
			continue
		}

		linksExpected := map[string]memory.Link{
			"abc": memory.Link{
				"mlink": "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V",
			},
		}
		linksActual := memory.Links(n)
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

func TestJsonStream(t *testing.T) {
	a := assrt.NewAssert(t)
	t.Logf("Reading json.testfile")
	json, err := Decode(bytes.NewReader(codedFiles["json.testfile"]))
	a.MustNil(err)

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
}

func TestCborStream(t *testing.T) {
	a := assrt.NewAssert(t)
	t.Logf("Reading cbor.testfile")
	cbor, err := Decode(bytes.NewReader(codedFiles["cbor.testfile"]))
	a.MustNil(err)

	readertest.CheckReader(t, cbor, []readertest.Callback{
		readertest.Callback{[]interface{}{}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{}, reader.TokenKey, "abc"},
		readertest.Callback{[]interface{}{"abc"}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{"abc"}, reader.TokenKey, "mlink"},
		readertest.Callback{[]interface{}{"abc", "mlink"}, reader.TokenValue, "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V"},
		readertest.Callback{[]interface{}{"abc"}, reader.TokenEndNode, nil},
		readertest.Callback{[]interface{}{}, reader.TokenKey, "@codec"},
		readertest.Callback{[]interface{}{"@codec"}, reader.TokenValue, "/json"},
		readertest.Callback{[]interface{}{}, reader.TokenEndNode, nil},
	})
}

func TestPbStream(t *testing.T) {
	a := assrt.NewAssert(t)
	t.Logf("Reading protobuf.testfile")
	pb, err := Decode(bytes.NewReader(codedFiles["protobuf.testfile"]))
	a.MustNil(err)

	readertest.CheckReader(t, pb, []readertest.Callback{
		readertest.Callback{[]interface{}{}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{}, reader.TokenKey, "data"},
		readertest.Callback{[]interface{}{"data"}, reader.TokenValue, []byte{0x08, 0x01}},
		readertest.Callback{[]interface{}{}, reader.TokenKey, "named-links"},
		readertest.Callback{[]interface{}{"named-links"}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{"named-links"}, reader.TokenKey, "a"},
		readertest.Callback{[]interface{}{"named-links", "a"}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{"named-links", "a"}, reader.TokenKey, "link"},
		readertest.Callback{[]interface{}{"named-links", "a", "link"}, reader.TokenValue, "Qmbvkmk9LFsGneteXk3G7YLqtLVME566ho6ibaQZZVHaC9"},
		readertest.Callback{[]interface{}{"named-links", "a"}, reader.TokenKey, "name"},
		readertest.Callback{[]interface{}{"named-links", "a", "name"}, reader.TokenValue, "a"},
		readertest.Callback{[]interface{}{"named-links", "a"}, reader.TokenKey, "size"},
		readertest.Callback{[]interface{}{"named-links", "a", "size"}, reader.TokenValue, uint64(10)},
		readertest.Callback{[]interface{}{"named-links", "a"}, reader.TokenEndNode, nil},
		readertest.Callback{[]interface{}{"named-links"}, reader.TokenKey, "b"},
		readertest.Callback{[]interface{}{"named-links", "b"}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{"named-links", "b"}, reader.TokenKey, "link"},
		readertest.Callback{[]interface{}{"named-links", "b", "link"}, reader.TokenValue, "QmR9pC5uCF3UExca8RSrCVL8eKv7nHMpATzbEQkAHpXmVM"},
		readertest.Callback{[]interface{}{"named-links", "b"}, reader.TokenKey, "name"},
		readertest.Callback{[]interface{}{"named-links", "b", "name"}, reader.TokenValue, "b"},
		readertest.Callback{[]interface{}{"named-links", "b"}, reader.TokenKey, "size"},
		readertest.Callback{[]interface{}{"named-links", "b", "size"}, reader.TokenValue, uint64(10)},
		readertest.Callback{[]interface{}{"named-links", "b"}, reader.TokenEndNode, nil},
		readertest.Callback{[]interface{}{"named-links"}, reader.TokenEndNode, nil},
		readertest.Callback{[]interface{}{}, reader.TokenKey, "ordered-links"},
		readertest.Callback{[]interface{}{"ordered-links"}, reader.TokenArray, nil},
		readertest.Callback{[]interface{}{"ordered-links"}, reader.TokenIndex, 0},
		readertest.Callback{[]interface{}{"ordered-links", 0}, reader.TokenValue, "a"},
		readertest.Callback{[]interface{}{"ordered-links"}, reader.TokenIndex, 1},
		readertest.Callback{[]interface{}{"ordered-links", 1}, reader.TokenValue, "b"},
		readertest.Callback{[]interface{}{"ordered-links"}, reader.TokenEndArray, nil},
		readertest.Callback{[]interface{}{}, reader.TokenEndNode, nil},
	})
}
