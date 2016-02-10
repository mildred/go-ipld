package cbor

import (
	"bytes"
	"testing"

	links "github.com/ipfs/go-ipld/links"
	memory "github.com/ipfs/go-ipld/memory"
	reader "github.com/ipfs/go-ipld/stream"
	readertest "github.com/ipfs/go-ipld/stream/test"
	multiaddr "github.com/jbenet/go-multiaddr"
	cbor "github.com/whyrusleeping/cbor/go"
)

func TestLinksStringEmptyMeta(t *testing.T) {
	var buf bytes.Buffer

	node := memory.Node{
		links.LinkKey: "#/foo/bar",
	}

	err := RawEncode(&buf, node, true)
	if err != nil {
		t.Error(err)
	}

	var expected []byte
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeTag, TagIPLDLink, nil)...)
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeText, uint64(len("#/foo/bar")), nil)...)
	expected = append(expected, []byte("#/foo/bar")...)

	if !bytes.Equal(expected, buf.Bytes()) {
		t.Error("Incorrect encoding")
		t.Logf("Expected: %v", expected)
		t.Logf("Actual:   %v", buf.Bytes())
	}

	cbor, err := NewCBORDecoder(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Error(err)
	}

	readertest.CheckReader(t, cbor, []readertest.Callback{
		readertest.Callback{[]interface{}{}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{}, reader.TokenKey, links.LinkKey},
		readertest.Callback{[]interface{}{links.LinkKey}, reader.TokenValue, "#/foo/bar"},
		readertest.Callback{[]interface{}{}, reader.TokenEndNode, nil},
	})
}

func TestLinksStringNonEmptyMetaCheckOrdering(t *testing.T) {
	var buf bytes.Buffer

	node := memory.Node{
		"size":        55,
		"00":          11,          // should be encoded first in the map (0 is before s and @ and is smaller)
		links.LinkKey: "#/foo/bar", // should come first, this is a link
	}

	err := RawEncode(&buf, node, true)
	if err != nil {
		t.Error(err)
	}

	var expected []byte
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeTag, TagIPLDLink, nil)...)
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeArray, 2, nil)...)
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeText, uint64(len("#/foo/bar")), nil)...)
	expected = append(expected, []byte("#/foo/bar")...)
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeMap, 2, nil)...)
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeText, uint64(len("00")), nil)...)
	expected = append(expected, []byte("00")...)
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeUint, 11, nil)...)
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeText, uint64(len("size")), nil)...)
	expected = append(expected, []byte("size")...)
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeUint, 55, nil)...)

	if !bytes.Equal(expected, buf.Bytes()) {
		t.Error("Incorrect encoding")
		t.Logf("Expected: %v", expected)
		t.Logf("Actual:   %v", buf.Bytes())
	}

	cbor, err := NewCBORDecoder(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Error(err)
	}

	readertest.CheckReader(t, cbor, []readertest.Callback{
		readertest.Callback{[]interface{}{}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{}, reader.TokenKey, links.LinkKey},
		readertest.Callback{[]interface{}{links.LinkKey}, reader.TokenValue, "#/foo/bar"},
		readertest.Callback{[]interface{}{}, reader.TokenKey, "00"},
		readertest.Callback{[]interface{}{"00"}, reader.TokenValue, uint64(11)},
		readertest.Callback{[]interface{}{}, reader.TokenKey, "size"},
		readertest.Callback{[]interface{}{"size"}, reader.TokenValue, uint64(55)},
		readertest.Callback{[]interface{}{}, reader.TokenEndNode, nil},
	})
}

func TestLinksMultiAddr(t *testing.T) {
	var buf bytes.Buffer

	ma, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/udp/1234")
	if err != nil {
		t.Error(err)
		return
	}

	node := memory.Node{
		links.LinkKey: ma.String(),
	}

	err = RawEncode(&buf, node, true)
	if err != nil {
		t.Error(err)
	}

	var expected []byte
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeTag, TagIPLDLink, nil)...)
	expected = append(expected, cbor.EncodeInt(cbor.MajorTypeBytes, uint64(len(ma.Bytes())), nil)...)
	expected = append(expected, ma.Bytes()...)

	if !bytes.Equal(expected, buf.Bytes()) {
		t.Error("Incorrect encoding")
		t.Logf("Expected: %v", expected)
		t.Logf("Actual:   %v", buf.Bytes())
	}

	cbor, err := NewCBORDecoder(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Error(err)
	}

	readertest.CheckReader(t, cbor, []readertest.Callback{
		readertest.Callback{[]interface{}{}, reader.TokenNode, nil},
		readertest.Callback{[]interface{}{}, reader.TokenKey, links.LinkKey},
		readertest.Callback{[]interface{}{links.LinkKey}, reader.TokenValue, ma.String()},
		readertest.Callback{[]interface{}{}, reader.TokenEndNode, nil},
	})
}
