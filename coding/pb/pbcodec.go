package ipldpb

import (
	"errors"
	"fmt"
	"io"

	mc "github.com/jbenet/go-multicodec"
	mcproto "github.com/jbenet/go-multicodec/protobuf"

	memory "github.com/ipfs/go-ipld/memory"
	base58 "github.com/jbenet/go-base58"
)

var Header []byte
var HeaderPath string

var (
	errInvalidData = fmt.Errorf("invalid merkledag v1 protobuf, Data not bytes")
	errInvalidLink = fmt.Errorf("invalid merkledag v1 protobuf, invalid Links")
)

func init() {
	HeaderPath = "/mdagv1"
	Header = mc.Header([]byte(HeaderPath))
}

type codec struct {
	pbc mc.Multicodec
}

func Multicodec() mc.Multicodec {
	var n *PBNode
	return &codec{mcproto.Multicodec(n)}
}

func (c *codec) Encoder(w io.Writer) mc.Encoder {
	return &encoder{w: w, c: c, pbe: c.pbc.Encoder(w)}
}

func (c *codec) Decoder(r io.Reader) mc.Decoder {
	return &decoder{r: r, c: c, pbd: c.pbc.Decoder(r)}
}

func (c *codec) Header() []byte {
	return Header
}

type encoder struct {
	w   io.Writer
	c   *codec
	pbe mc.Encoder
}

type decoder struct {
	r   io.Reader
	c   *codec
	pbd mc.Decoder
}

func (c *encoder) Encode(v interface{}) error {
	nv, ok := v.(*memory.Node)
	if !ok {
		return errors.New("must encode *memory.Node")
	}

	if _, err := c.w.Write(c.c.Header()); err != nil {
		return err
	}

	n, err := ld2pbNode(nv)
	if err != nil {
		return err
	}

	return c.pbe.Encode(n)
}

func (c *decoder) Decode(v interface{}) error {
	nv, ok := v.(*memory.Node)
	if !ok {
		return errors.New("must decode to *memory.Node")
	}

	if err := mc.ConsumeHeader(c.r, c.c.Header()); err != nil {
		return err
	}

	var pbn PBNode
	if err := c.pbd.Decode(&pbn); err != nil {
		return err
	}

	pb2ldNode(&pbn, nv)
	return nil
}

func ld2pbNode(in *memory.Node) (*PBNode, error) {
	n := *in
	var pbn PBNode
	var ordered_links []interface{}
	var named_links memory.Node

	if data, hasdata := n["data"]; hasdata {
		data, ok := data.([]byte)
		if !ok {
			return nil, errInvalidData
		}
		pbn.Data = data
	}

	if ordered_links_node, ok := n["ordered-links"]; ok {
		var ok bool
		ordered_links, ok = ordered_links_node.([]interface{})
		if !ok {
			return nil, errInvalidData
		}
	} else {
		return nil, errInvalidData
	}

	if named_links_node, ok := n["named-links"]; ok {
		var ok bool
		named_links, ok = named_links_node.(memory.Node)
		if !ok {
			return nil, errInvalidData
		}
	} else {
		return nil, errInvalidData
	}

	for n, link := range ordered_links {
		var pblink *PBLink
		var linkname string
		switch link.(type) {
		case string:
			linkname = link.(string)
			linknode := named_links[linkname].(memory.Node)
			if linknode != nil {
				pblink = ld2pbLink(linknode)
			}
		case memory.Node:
			pblink = ld2pbLink(link.(memory.Node))
			linkname = link.(memory.Node)["name"].(string)
		}
		if pblink == nil {
			return nil, fmt.Errorf("%s (#%d %s)", errInvalidLink, n, linkname)
		}
		pbn.Links = append(pbn.Links, pblink)
	}
	return &pbn, nil
}

func pb2ldNode(pbn *PBNode, in *memory.Node) {
	var ordered_links []interface{}
	var named_links memory.Node = make(memory.Node)

	for _, link := range pbn.Links {
		if _, exists := named_links[link.GetName()]; exists {
			ordered_links = append(ordered_links, pb2ldLink(link))
		} else {
			ordered_links = append(ordered_links, link.GetName())
			named_links[link.GetName()] = pb2ldLink(link)
		}
	}

	(*in)["data"] = pbn.GetData()
	(*in)["named-links"] = named_links
	(*in)["ordered-links"] = ordered_links
}

func pb2ldLink(pbl *PBLink) (link memory.Node) {
	defer func() {
		if recover() != nil {
			link = nil
		}
	}()

	link = make(memory.Node)
	link["link"] = base58.Encode(pbl.Hash)
	link["name"] = *pbl.Name
	link["size"] = uint64(*pbl.Tsize)
	return link
}

func ld2pbLink(link memory.Node) (pbl *PBLink) {
	defer func() {
		if recover() != nil {
			pbl = nil
		}
	}()

	hash := base58.Decode(link["hash"].(string))
	name := link["name"].(string)
	size := link["size"].(uint64)

	pbl = &PBLink{}
	pbl.Hash = hash
	pbl.Name = &name
	pbl.Tsize = &size
	return pbl
}
