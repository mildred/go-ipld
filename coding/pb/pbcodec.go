package ipldpb

import (
	"fmt"
	"io"

	links "github.com/ipfs/go-ipld/links"
	memory "github.com/ipfs/go-ipld/memory"
	base58 "github.com/jbenet/go-base58"
	msgio "github.com/jbenet/go-msgio"
	mc "github.com/jbenet/go-multicodec"
)

const HeaderPath = "/mdagv1"
const MsgIOHeaderPath = "/protobuf/msgio"

var Header []byte
var MsgIOHeader []byte

var errInvalidLink = fmt.Errorf("invalid merkledag v1 protobuf, invalid Links")

func init() {
	Header = mc.Header([]byte(HeaderPath))
	MsgIOHeader = mc.Header([]byte(MsgIOHeaderPath))
}

func Decode(r io.Reader) (memory.Node, error) {
	err := mc.ConsumeHeader(r, MsgIOHeader)
	if err != nil {
		return nil, err
	}

	length, err := msgio.ReadLen(r, nil)
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	_, err = io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}

	return RawDecode(data)
}

func RawDecode(data []byte) (memory.Node, error) {
	var pbn *PBNode = new(PBNode)

	err := pbn.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	n := make(memory.Node)
	pb2ldNode(pbn, &n)

	return n, err
}

func Encode(w io.Writer, n memory.Node, strict bool) error {
	_, err := w.Write(MsgIOHeader)
	if err != nil {
		return err
	}

	data, err := RawEncode(n, strict)
	if err != nil {
		return err
	}

	msgio.WriteLen(w, len(data))
	_, err = w.Write(data)
	return err
}

func RawEncode(n memory.Node, strict bool) ([]byte, error) {
	pbn, err := ld2pbNode(&n, strict)
	if err != nil {
		return nil, err
	}

	return pbn.Marshal()
}

func ld2pbNode(in *memory.Node, strict bool) (*PBNode, error) {
	n := *in
	var pbn PBNode
	size := 0

	if data, hasdata := n["data"]; hasdata {
		size++
		data, ok := data.([]byte)
		if !ok {
			return nil, fmt.Errorf("Invalid merkledag v1 protobuf: data is of incorect type")
		}
		pbn.Data = data
	} else if strict {
		return nil, fmt.Errorf("Invalid merkledag v1 protobuf: no data")
	}

	if links, ok := n["links"].([]interface{}); ok && links != nil {
		size++
		for _, link := range links {
			l, ok := link.(memory.Node)
			if !ok {
				return nil, fmt.Errorf("Invalid merkledag v1 protobuf: link is of incorect type")
			}
			pblink, err := ld2pbLink(l, strict)
			if err != nil {
				return nil, err
			}
			pbn.Links = append(pbn.Links, pblink)
		}
	} else if strict {
		return nil, fmt.Errorf("Invalid merkledag v1 protobuf: no links")
	}

	if strict && len(n) != size {
		return nil, fmt.Errorf("Invalid merkledag v1 protobuf: node contains extra fields (%d)", len(n)-size)
	}

	return &pbn, nil
}

func pb2ldNode(pbn *PBNode, in *memory.Node) {
	var ordered_links []interface{}

	for _, link := range pbn.Links {
		ordered_links = append(ordered_links, pb2ldLink(link))
	}

	(*in)["data"] = pbn.GetData()
	(*in)["links"] = ordered_links
}

func pb2ldLink(pbl *PBLink) (link memory.Node) {
	defer func() {
		if recover() != nil {
			link = nil
		}
	}()

	link = make(memory.Node)
	link[links.LinkKey] = base58.Encode(pbl.Hash)
	link["name"] = *pbl.Name
	link["size"] = uint64(*pbl.Tsize)
	return link
}

func ld2pbLink(link memory.Node, strict bool) (pbl *PBLink, err error) {
	length := 0
	pbl = &PBLink{}

	if hash, ok := link[links.LinkKey].(string); ok {
		length++
		pbl.Hash = base58.Decode(hash)
		if strict && base58.Encode(pbl.Hash) != hash {
			return nil, errInvalidLink
		}
	}

	if name, ok := link["name"].(string); ok {
		length++
		pbl.Name = &name
	}

	if size, ok := link["size"].(uint64); ok {
		length++
		pbl.Tsize = &size
	}

	if strict && len(link) != length {
		return nil, fmt.Errorf("Invalid merkledag v1 protobuf: link contains %d fields instead of %d", len(link), length)
	}

	return pbl, err
}
