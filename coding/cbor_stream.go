package coding

import (
	"fmt"
	"io"
	"math/big"

	reader "github.com/ipfs/go-ipld"
	cbor "github.com/whyrusleeping/cbor/go"
)

type CBORDecoder struct {
	r   io.Reader
	pos int64
}

type cborParser struct {
	reader.BaseReader
	decoder *cbor.Decoder
}

func NewCBORDecoder(r io.Reader) (*CBORDecoder, error) {
	s := r.(io.Seeker)
	if s == nil {
		return &CBORDecoder{r, -1}, nil
	} else {
		offset, err := s.Seek(0, 1)
		if err != nil {
			return nil, err
		}
		return &CBORDecoder{r, offset}, nil
	}
}

func (d *CBORDecoder) Read(cb reader.ReadFun) error {
	if d.pos == -2 {
		return ErrAlreadyRead
	} else if d.pos == -1 {
		d.pos = -2
	} else {
		newoffset, err := d.r.(io.Seeker).Seek(d.pos, 0)
		if err != nil {
			return err
		} else if newoffset != d.pos {
			return fmt.Errorf("Failed to seek to position %d", d.pos)
		}
	}
	dec := cbor.NewDecoder(d.r)
	return dec.DecodeAny(&cborParser{reader.CreateBaseReader(cb), dec})
}

func (p *cborParser) Prepare() error {
	return nil
}

func (p *cborParser) SetBytes(buf []byte) error {
	err := p.ExecCallback(reader.TokenValue, buf)
	p.Descope()
	return err
}

func (p *cborParser) SetUint(i uint64) error {
	err := p.ExecCallback(reader.TokenValue, i)
	p.Descope()
	return err
}

func (p *cborParser) SetInt(i int64) error {
	err := p.ExecCallback(reader.TokenValue, i)
	p.Descope()
	return err
}

func (p *cborParser) SetFloat32(f float32) error {
	err := p.ExecCallback(reader.TokenValue, f)
	p.Descope()
	return err
}

func (p *cborParser) SetFloat64(f float64) error {
	err := p.ExecCallback(reader.TokenValue, f)
	p.Descope()
	return err
}

func (p *cborParser) SetBignum(i *big.Int) error {
	err := p.ExecCallback(reader.TokenValue, i)
	p.Descope()
	return err
}

func (p *cborParser) SetNil() error {
	err := p.ExecCallback(reader.TokenValue, nil)
	p.Descope()
	return err
}

func (p *cborParser) SetBool(b bool) error {
	err := p.ExecCallback(reader.TokenValue, b)
	p.Descope()
	return err
}

func (p *cborParser) SetString(s string) error {
	err := p.ExecCallback(reader.TokenValue, s)
	p.Descope()
	return err
}

func (p *cborParser) CreateMap() (cbor.DecodeValueMap, error) {
	return p, p.ExecCallback(reader.TokenNode, nil)
}

func (p *cborParser) CreateMapKey() (cbor.DecodeValue, error) {
	return cbor.NewMemoryValue(""), nil
}

func (p *cborParser) CreateMapValue(key cbor.DecodeValue) (cbor.DecodeValue, error) {
	err := p.ExecCallback(reader.TokenKey, key.(*cbor.MemoryValue).Value)
	p.Descope()
	p.PushPath(key.(*cbor.MemoryValue).Value)
	return p, err
}

func (p *cborParser) SetMap(key, val cbor.DecodeValue) error {
	p.PopPath()
	return nil
}

func (p *cborParser) EndMap() error {
	err := p.ExecCallback(reader.TokenEndNode, nil)
	p.Descope()
	p.Descope()
	return err
}

func (p *cborParser) CreateArray(len int) (cbor.DecodeValueArray, error) {
	return p, p.ExecCallback(reader.TokenArray, nil)
}

func (p *cborParser) GetArrayValue(index uint64) (cbor.DecodeValue, error) {
	err := p.ExecCallback(reader.TokenIndex, index)
	p.Descope()
	p.PushPath(index)
	return p, err
}

func (p *cborParser) AppendArray(val cbor.DecodeValue) error {
	p.PopPath()
	return nil
}

func (p *cborParser) EndArray() error {
	err := p.ExecCallback(reader.TokenEndArray, nil)
	p.Descope()
	p.Descope()
	return err
}

func (p *cborParser) CreateTag(tag uint64, decoder cbor.TagDecoder) (cbor.DecodeValue, interface{}, error) {
	return p, nil, nil
}

func (p *cborParser) SetTag(tag uint64, decval cbor.DecodeValue, decoder cbor.TagDecoder, val interface{}) error {
	return nil
}
