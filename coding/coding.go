package ipfsld

import (
	"fmt"
	"io"

	mc "github.com/jbenet/go-multicodec"
	mccbor "github.com/jbenet/go-multicodec/cbor"
	mcjson "github.com/jbenet/go-multicodec/json"
	mcmux "github.com/jbenet/go-multicodec/mux"

	ipld "github.com/ipfs/go-ipld"
	pb "github.com/ipfs/go-ipld/coding/pb"
	reader "github.com/ipfs/go-ipld/reader"
)

var StreamCodecs map[string]func(io.Reader) (reader.NodeReader, error) = map[string]func(io.Reader) (reader.NodeReader, error){
	mcjson.HeaderPath: func(r io.Reader) (reader.NodeReader, error) {
		return &JSONDecoder{r}, nil
	},
	mccbor.HeaderPath: func(r io.Reader) (reader.NodeReader, error) {
		return &CBORDecoder{r}, nil
	},
}

// defaultCodec is the default applied if user does not specify a codec.
// Most new objects will never specify a codec. We track the codecs with
// the object so that multiple people using the same object will continue
// to marshal using the same codec. the only reason this is important is
// that the hashes must be the same.
var defaultCodec string

var muxCodec *mcmux.Multicodec

func init() {
	// by default, always encode things as cbor
	defaultCodec = string(mc.HeaderPath(mccbor.Header))
	muxCodec = mcmux.MuxMulticodec([]mc.Multicodec{
		CborMulticodec(),
		JsonMulticodec(),
		pb.Multicodec(),
	}, selectCodec)
	StreamCodecs = map[string]func(io.Reader) (reader.NodeReader, error){
		mcjson.HeaderPath: func(r io.Reader) (reader.NodeReader, error) {
			return &JSONDecoder{r}, nil
		},
		mccbor.HeaderPath: func(r io.Reader) (reader.NodeReader, error) {
			return &CBORDecoder{r}, nil
		},
	}
}

// Multicodec returns a muxing codec that marshals to
// whatever codec makes sense depending on what information
// the IPLD object itself carries
func Multicodec() mc.Multicodec {
	return muxCodec
}

func selectCodec(v interface{}, codecs []mc.Multicodec) mc.Multicodec {
	vn, ok := v.(*ipld.Node)
	if !ok {
		return nil
	}

	codecKey, err := codecKey(*vn)
	if err != nil {
		return nil
	}

	for _, c := range codecs {
		if codecKey == string(mc.HeaderPath(c.Header())) {
			return c
		}
	}

	return nil // no codec
}

func codecKey(n ipld.Node) (string, error) {
	chdr, ok := (n)[ipld.CodecKey]
	if !ok {
		// if no codec is defined, use our default codec
		chdr = defaultCodec
		if pb.IsOldProtobufNode(n) {
			chdr = string(pb.Header)
		}
	}

	chdrs, ok := chdr.(string)
	if !ok {
		// if chdr is not a string, cannot read codec.
		return "", mc.ErrType
	}

	return chdrs, nil
}

func Decode(r io.Reader) (reader.NodeReader, error) {
	if err := mc.ConsumeHeader(r, mcmux.Header); err != nil {
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
		return nil, fmt.Errorf("no codec for %s", hdr)
	}
	return fun(r)
}
