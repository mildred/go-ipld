package coding

import (
	"fmt"
	"io"

	mc "github.com/jbenet/go-multicodec"
	mccbor "github.com/jbenet/go-multicodec/cbor"
	mcjson "github.com/jbenet/go-multicodec/json"
	mcmux "github.com/jbenet/go-multicodec/mux"
	mcproto "github.com/jbenet/go-multicodec/protobuf"

	reader "github.com/ipfs/go-ipld"
	pb "github.com/ipfs/go-ipld/coding/pb"
	memory "github.com/ipfs/go-ipld/memory"
)

var Header []byte

const (
	HeaderPath   = "/mdagv1"
	ProtobufPath = "/protobuf/msgio"
)

var StreamCodecs map[string]func(io.Reader) (reader.NodeReader, error)

// defaultCodec is the default applied if user does not specify a codec.
// Most new objects will never specify a codec. We track the codecs with
// the object so that multiple people using the same object will continue
// to marshal using the same codec. the only reason this is important is
// that the hashes must be the same.
var defaultCodec string

var muxCodec *mcmux.Multicodec

var ErrAlreadyRead error = fmt.Errorf("Stream already read: unable to read it a second time")

func init() {
	Header = mc.Header([]byte(HeaderPath))

	// by default, always encode things as cbor
	defaultCodec = string(mc.HeaderPath(mccbor.Header))

	muxCodec = mcmux.MuxMulticodec([]mc.Multicodec{
		CborMulticodec(),
		JsonMulticodec(),
		pb.Multicodec(),
	}, selectCodec)

	StreamCodecs = map[string]func(io.Reader) (reader.NodeReader, error){
		mcjson.HeaderPath: func(r io.Reader) (reader.NodeReader, error) {
			return NewJSONDecoder(r)
		},
		mccbor.HeaderPath: func(r io.Reader) (reader.NodeReader, error) {
			return NewCBORDecoder(r)
		},
		ProtobufPath: DecodeLegacyProtobuf,
	}
}

// Multicodec returns a muxing codec that marshals to
// whatever codec makes sense depending on what information
// the IPLD object itself carries
func Multicodec() mc.Multicodec {
	return muxCodec
}

func selectCodec(v interface{}, codecs []mc.Multicodec) mc.Multicodec {
	return nil // no codec
}

func Decode(r io.Reader) (reader.NodeReader, error) {
	// get multicodec first header, should be mcmux.Header
	err := mc.ConsumeHeader(r, Header)
	if err != nil {
		return nil, err
	}

	// get next header, to select codec
	hdr, err := mc.ReadHeader(r)
	if err != nil {
		return nil, err
	}

	hdrPath := string(mc.HeaderPath(hdr))

	fun, ok := StreamCodecs[hdrPath]
	if !ok {
		return nil, fmt.Errorf("no codec for %s", hdrPath)
	}
	return fun(r)
}

func DecodeLegacyProtobuf(r io.Reader) (reader.NodeReader, error) {
	var node memory.Node
	r = mc.WrapHeaderReader(mcproto.HeaderMsgio, r)
	r = mc.WrapHeaderReader(pb.Header, r)
	err := pb.Multicodec().Decoder(r).Decode(&node)
	return node, err
}
