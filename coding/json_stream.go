package ipfsld

import (
	"encoding/json"
	"fmt"
	"io"

	reader "github.com/ipfs/go-ipld/reader"
)

type JSONDecoder struct {
	r io.Reader
}

type jsonParser struct {
	reader.BaseReader
	decoder *json.Decoder
}

func (d *JSONDecoder) Read(cb reader.ReadFun) error {
	jsonParser := &jsonParser{
		reader.CreateBaseReader(cb),
		json.NewDecoder(d.r),
	}
	err := jsonParser.readValue()
	if err == reader.NodeReadAbort {
		err = nil
	}
	return err
}

func (p *jsonParser) readValue() error {
	token, err := p.decoder.Token()
	if err != nil {
		return err
	}
	//log.Printf("JSON: read token value %#v %T", token, token)
	if delim, ok := token.(json.Delim); ok {
		switch delim {
		case '{':
			err = p.ExecCallback(reader.TokenNode, nil)
			if err != nil {
				p.Descope()
				return err
			}
			err = p.readNode()
			if err != nil {
				p.Descope()
				return err
			}
			err = p.ExecCallback(reader.TokenEndNode, nil)
			p.Descope()
			p.Descope()
			return err
			break
		case '[':
			err = p.ExecCallback(reader.TokenArray, nil)
			if err != nil {
				p.Descope()
				return err
			}
			err = p.readArray()
			if err != nil {
				p.Descope()
				return err
			}
			err = p.ExecCallback(reader.TokenEndArray, nil)
			p.Descope()
			p.Descope()
			return err
			break
		default:
			return fmt.Errorf("JSON: unexpected delimiter token %#v", token.(json.Delim))
		}
	} else {
		switch token.(type) {

		case json.Number:
			intValue, err1 := token.(json.Number).Int64()
			if err1 != nil {
				token = intValue
			} else {
				token, err = token.(json.Number).Float64()
				if err != nil {
					return fmt.Errorf("JSON: failed to convert %v to float64: %v", token.(json.Number), err)
				}
			}
		case float64:
			if sintValue := int(token.(float64)); token.(float64) == float64(sintValue) {
				token = sintValue
			} else if intValue := int64(token.(float64)); token.(float64) == float64(intValue) {
				token = intValue
			} else if uintValue := uint64(token.(float64)); token.(float64) == float64(uintValue) {
				token = uintValue
			}
		case string:
		case bool:
		case nil:
		default:
			return fmt.Errorf("JSON: Unexpected token %#v", token)
		}
		err := p.ExecCallback(reader.TokenValue, token)
		p.Descope()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *jsonParser) readNode() error {
	for {
		token, err := p.decoder.Token()
		if err != nil {
			return err
		}
		//log.Printf("JSON: read token node  %#v %T", token, token)

		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}

		strValue, isStr := token.(string)
		if !isStr {
			return fmt.Errorf("JSON: expect string for object key: got %#v", token)
		}
		err = p.ExecCallback(reader.TokenKey, strValue)
		p.Descope()
		if err != nil {
			return err
		}

		p.PushPath(strValue)
		err = p.readValue()
		p.PopPath()
		if err != nil {
			return err
		}
	}
}

func (p *jsonParser) readArray() error {
	var index uint64 = 0
	for {
		token, err := p.decoder.Token()
		if err != nil {
			return err
		}
		//log.Printf("JSON: read token array %#v %T", token, token)

		if delim, ok := token.(json.Delim); ok && delim == ']' {
			return nil
		}

		p.PushPath(index)
		err = p.readValue()
		p.PopPath()
		if err != nil {
			return err
		}

		index++
	}
}
