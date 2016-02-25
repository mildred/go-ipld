package coding

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	links "github.com/ipfs/go-ipld/links"
	memory "github.com/ipfs/go-ipld/memory"
	reader "github.com/ipfs/go-ipld/stream"
	readertest "github.com/ipfs/go-ipld/stream/test"
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

// Test decoding and encoding a json and cbor file
func TestCodecsEncodeDecode(t *testing.T) {
	for fname, testfile := range codedFiles {

		r, err := DecodeBytes(testfile)
		if err != nil {
			t.Error(err)
			continue
		}

		var codec Codec
		switch fname {
		case "json.testfile":
			codec = CodecJSON
		case "cbor.testfile":
			codec = CodecCBOR
		case "protobuf.testfile":
			codec = CodecProtobuf
		default:
			panic("should not arrive here")
		}

		t.Logf("Decoded %s: %#v", fname, r)

		n, err := memory.NewNodeFrom(r)
		if err != nil {
			t.Error(err)
			continue
		}

		t.Logf("In memory %s: %#v", fname, n)

		outData, err := EncodeBytes(codec, n)
		if err != nil {
			t.Error(err)
			continue
		}

		if !bytes.Equal(outData, testfile) {
			t.Errorf("%s: encoded is not the same as original", fname)
			t.Log(n)
			t.Log(testfile)
			t.Log(string(testfile))
			t.Log(outData)
			t.Log(string(outData))
			f, err := os.Create(fname + ".error")
			if err != nil {
				t.Error(err)
			} else {
				defer f.Close()
				_, err := f.Write(outData)
				if err != nil {
					t.Error(err)
				}
			}
		}
	}
}

func TestJsonStream(t *testing.T) {
	a := assrt.NewAssert(t)
	t.Logf("Reading json.testfile")
	json, err := Decode(bytes.NewReader(codedFiles["json.testfile"]))
	a.MustNil(err)

	readertest.CheckReader(t, json, []readertest.Callback{
		readertest.Cb(readertest.Path(), reader.TokenNode, nil),
		readertest.Cb(readertest.Path(), reader.TokenKey, "@codec"),
		readertest.Cb(readertest.Path("@codec"), reader.TokenValue, "/json"),
		readertest.Cb(readertest.Path(), reader.TokenKey, "abc"),
		readertest.Cb(readertest.Path("abc"), reader.TokenNode, nil),
		readertest.Cb(readertest.Path("abc"), reader.TokenKey, "mlink"),
		readertest.Cb(readertest.Path("abc", "mlink"), reader.TokenValue, "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V"),
		readertest.Cb(readertest.Path("abc"), reader.TokenEndNode, nil),
		readertest.Cb(readertest.Path(), reader.TokenEndNode, nil),
	})
}

func TestJsonStreamSkip(t *testing.T) {
	a := assrt.NewAssert(t)
	t.Logf("Reading json.testfile")
	json, err := Decode(bytes.NewReader(codedFiles["json.testfile"]))
	a.MustNil(err)

	readertest.CheckReader(t, json, []readertest.Callback{
		readertest.Cb(readertest.Path(), reader.TokenNode, nil),
		readertest.Cb(readertest.Path(), reader.TokenKey, "@codec", reader.NodeReadSkip),
		readertest.Cb(readertest.Path(), reader.TokenKey, "abc"),
		readertest.Cb(readertest.Path("abc"), reader.TokenNode, nil),
		readertest.Cb(readertest.Path("abc"), reader.TokenKey, "mlink", reader.NodeReadAbort),
	})
}

func TestCborStream(t *testing.T) {
	a := assrt.NewAssert(t)
	t.Logf("Reading cbor.testfile")
	cbor, err := Decode(bytes.NewReader(codedFiles["cbor.testfile"]))
	a.MustNil(err)

	readertest.CheckReader(t, cbor, []readertest.Callback{
		readertest.Cb(readertest.Path(), reader.TokenNode, nil),
		readertest.Cb(readertest.Path(), reader.TokenKey, "abc"),
		readertest.Cb(readertest.Path("abc"), reader.TokenNode, nil),
		readertest.Cb(readertest.Path("abc"), reader.TokenKey, "mlink"),
		readertest.Cb(readertest.Path("abc", "mlink"), reader.TokenValue, "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V"),
		readertest.Cb(readertest.Path("abc"), reader.TokenEndNode, nil),
		readertest.Cb(readertest.Path(), reader.TokenKey, "@codec"),
		readertest.Cb(readertest.Path("@codec"), reader.TokenValue, "/json"),
		readertest.Cb(readertest.Path(), reader.TokenEndNode, nil),
	})
}

func TestPbStream(t *testing.T) {
	a := assrt.NewAssert(t)
	t.Logf("Reading protobuf.testfile")
	t.Logf("Bytes: %v", codedFiles["protobuf.testfile"])
	pb, err := Decode(bytes.NewReader(codedFiles["protobuf.testfile"]))
	a.MustNil(err)

	readertest.CheckReader(t, pb, []readertest.Callback{
		readertest.Cb(readertest.Path(), reader.TokenNode, nil),
		readertest.Cb(readertest.Path(), reader.TokenKey, "data"),
		readertest.Cb(readertest.Path("data"), reader.TokenValue, []byte{0x08, 0x01}),
		readertest.Cb(readertest.Path(), reader.TokenKey, "links"),
		readertest.Cb(readertest.Path("links"), reader.TokenArray, nil),
		readertest.Cb(readertest.Path("links"), reader.TokenIndex, 0),
		readertest.Cb(readertest.Path("links", 0), reader.TokenNode, nil),
		readertest.Cb(readertest.Path("links", 0), reader.TokenKey, links.LinkKey),
		readertest.Cb(readertest.Path("links", 0, links.LinkKey), reader.TokenValue, "Qmbvkmk9LFsGneteXk3G7YLqtLVME566ho6ibaQZZVHaC9"),
		readertest.Cb(readertest.Path("links", 0), reader.TokenKey, "name"),
		readertest.Cb(readertest.Path("links", 0, "name"), reader.TokenValue, "a"),
		readertest.Cb(readertest.Path("links", 0), reader.TokenKey, "size"),
		readertest.Cb(readertest.Path("links", 0, "size"), reader.TokenValue, uint64(10)),
		readertest.Cb(readertest.Path("links", 0), reader.TokenEndNode, nil),
		readertest.Cb(readertest.Path("links"), reader.TokenIndex, 1),
		readertest.Cb(readertest.Path("links", 1), reader.TokenNode, nil),
		readertest.Cb(readertest.Path("links", 1), reader.TokenKey, links.LinkKey),
		readertest.Cb(readertest.Path("links", 1, links.LinkKey), reader.TokenValue, "QmR9pC5uCF3UExca8RSrCVL8eKv7nHMpATzbEQkAHpXmVM"),
		readertest.Cb(readertest.Path("links", 1), reader.TokenKey, "name"),
		readertest.Cb(readertest.Path("links", 1, "name"), reader.TokenValue, "b"),
		readertest.Cb(readertest.Path("links", 1), reader.TokenKey, "size"),
		readertest.Cb(readertest.Path("links", 1, "size"), reader.TokenValue, uint64(10)),
		readertest.Cb(readertest.Path("links", 1), reader.TokenEndNode, nil),
		readertest.Cb(readertest.Path("links"), reader.TokenEndArray, nil),
		readertest.Cb(readertest.Path(), reader.TokenEndNode, nil),
	})
}
