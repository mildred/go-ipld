package ipfsld

import (
	"io"
	"log"
	"math/big"

	reader "github.com/ipfs/go-ipld/reader"
	cbor "github.com/whyrusleeping/cbor/go"
)

type CBORDecoder struct {
	r io.Reader
}

type cborParser struct {
	reader.BaseReader
	decoder *cbor.Decoder
}

func (d *CBORDecoder) Read(cb reader.ReadFun) error {
	dec := cbor.NewDecoder(d.r)
	return dec.DecodeAny(&cborParser{reader.CreateBaseReader(cb), dec})
}

func (p *cborParser) Prepare() error {
	log.Printf("Prepare")
	return nil
}

func (p *cborParser) SetBytes(buf []byte) error {
	log.Printf("SetBytes")
	err := p.ExecCallback(reader.TokenValue, buf)
	p.Descope()
	return err
}

func (p *cborParser) SetUint(i uint64) error {
	log.Printf("SetUint")
	err := p.ExecCallback(reader.TokenValue, i)
	p.Descope()
	return err
}

func (p *cborParser) SetInt(i int64) error {
	log.Printf("setint")
	err := p.ExecCallback(reader.TokenValue, i)
	p.Descope()
	return err
}

func (p *cborParser) SetFloat32(f float32) error {
	log.Printf("setfloat32")
	err := p.ExecCallback(reader.TokenValue, f)
	p.Descope()
	return err
}

func (p *cborParser) SetFloat64(f float64) error {
	log.Printf("setfloat64")
	err := p.ExecCallback(reader.TokenValue, f)
	p.Descope()
	return err
}

func (p *cborParser) SetBignum(i *big.Int) error {
	log.Printf("setbignum")
	err := p.ExecCallback(reader.TokenValue, i)
	p.Descope()
	return err
}

func (p *cborParser) SetNil() error {
	log.Printf("nil")
	err := p.ExecCallback(reader.TokenValue, nil)
	p.Descope()
	return err
}

func (p *cborParser) SetBool(b bool) error {
	log.Printf("setbool")
	err := p.ExecCallback(reader.TokenValue, b)
	p.Descope()
	return err
}

func (p *cborParser) SetString(s string) error {
	log.Printf("setstring")
	err := p.ExecCallback(reader.TokenValue, s)
	p.Descope()
	return err
}

func (p *cborParser) CreateMap() (cbor.DecodeValueMap, error) {
	log.Printf("createmap")
	return p, p.ExecCallback(reader.TokenNode, nil)
}

func (p *cborParser) CreateMapKey() (cbor.DecodeValue, error) {
	log.Printf("createmapkey")
	return cbor.NewMemoryValue(""), nil
}

func (p *cborParser) CreateMapValue(key cbor.DecodeValue) (cbor.DecodeValue, error) {
	log.Printf("createmapvalue")
	err := p.ExecCallback(reader.TokenKey, key.(*cbor.MemoryValue).Value)
	p.Descope()
	p.PushPath(key.(*cbor.MemoryValue).Value)
	return p, err
}

func (p *cborParser) SetMap(key, val cbor.DecodeValue) error {
	log.Printf("setmap")
	p.PopPath()
	return nil
}

func (p *cborParser) EndMap() error {
	log.Printf("endmap")
	err := p.ExecCallback(reader.TokenEndNode, nil)
	p.Descope()
	p.Descope()
	return err
}

func (p *cborParser) CreateArray(len int) (cbor.DecodeValueArray, error) {
	log.Printf("createarray")
	return p, p.ExecCallback(reader.TokenArray, nil)
}

func (p *cborParser) GetArrayValue(index uint64) (cbor.DecodeValue, error) {
	log.Printf("getarrvalue")
	err := p.ExecCallback(reader.TokenIndex, index)
	p.Descope()
	p.PushPath(index)
	return p, err
}

func (p *cborParser) AppendArray(val cbor.DecodeValue) error {
	log.Printf("appendarray")
	p.PopPath()
	return nil
}

func (p *cborParser) EndArray() error {
	log.Printf("endarray")
	err := p.ExecCallback(reader.TokenEndArray, nil)
	p.Descope()
	p.Descope()
	return err
}

func (p *cborParser) CreateTag(tag uint64, decoder cbor.TagDecoder) (cbor.DecodeValue, interface{}, error) {
	log.Printf("createtag")
	return p, nil, nil
}

func (p *cborParser) SetTag(tag uint64, decval cbor.DecodeValue, decoder cbor.TagDecoder, val interface{}) error {
	log.Printf("settag")
	return nil
}
