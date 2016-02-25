package ipld

import (
	"fmt"
	"reflect"

	links "github.com/ipfs/go-ipld/links"
	stream "github.com/ipfs/go-ipld/stream"
	mh "github.com/jbenet/go-multihash"
)

// Read the hash of an IPLD link.
func ReadLinkPath(value interface{}) (mh.Multihash, error) {
	svalue, has_value := value.(string)
	var h mh.Multihash
	if has_value {
		return mh.FromB58String(svalue)
	}
	return h, fmt.Errorf("Could not get multihash for %#v", value)
}

func setValue(f reflect.Value, it *stream.NodeIterator) (err error) {
	defer func() {
		ex := recover()
		if e, ok := ex.(error); ex != nil && ok {
			err = e
		} else if ex != nil {
			err = fmt.Errorf("%#v", ex)
		}
	}()

	v := reflect.ValueOf(it.Value)
	if f.Type().AssignableTo(reflect.TypeOf(it.Value)) {
		f.Set(v)
	} else if ui, ok := it.ToUint(); ok && f.Type().AssignableTo(reflect.TypeOf(ui)) {
		f.SetUint(ui)
	} else if i, ok := it.ToInt(); ok && f.Type().AssignableTo(reflect.TypeOf(i)) {
		f.SetInt(i)
	} else if ff, ok := it.ToFloat(); ok && f.Type().AssignableTo(reflect.TypeOf(ff)) {
		f.SetFloat(ff)
	} else if b, ok := it.ToBytes(); ok && f.Type().AssignableTo(reflect.TypeOf(b)) {
		f.SetBytes(b)
	}
	return nil
}

type fields struct {
	hash  []int
	link  []int
	name  []int
	spath []int
	ipath []int
	attrs map[string][]int
}

var mhType reflect.Type = reflect.TypeOf(mh.Multihash{})

func (f *fields) mapFields(lt reflect.Type, path []int) {
	if f.attrs == nil {
		f.attrs = make(map[string][]int)
	}
	for i := 0; i < lt.NumField(); i++ {
		lf := lt.Field(i)
		curPath := append(path, i)

		if lf.PkgPath != "" {
			// unexported
			continue
		}

		tag := lf.Tag.Get("ipld")

		if tag == "multihash" && f.hash == nil && lf.Type.AssignableTo(mhType) {
			f.hash = curPath
		} else if len(tag) > 4 && tag[:4] == "key:" {
			f.attrs[tag[4:]] = curPath
		} else if f.link == nil && tag == "link" && lf.Type.Kind() == reflect.String {
			f.link = curPath
		} else if f.name == nil && tag == "name" && lf.Type.Kind() == reflect.String {
			f.name = curPath
		} else if f.spath == nil && tag == "path" && lf.Type.Kind() == reflect.Slice && lf.Type.Elem().Kind() == reflect.String {
			f.spath = curPath
		} else if f.ipath == nil && tag == "path" && lf.Type.Kind() == reflect.Slice && lf.Type.Elem().Kind() == reflect.Interface && lf.Type.Elem().NumMethod() == 0 {
			f.ipath = curPath
		} else if lf.Anonymous && lf.Type.Kind() == reflect.Struct {
			f.mapFields(lf.Type, curPath)
		}
	}
}

// Read the links of r. links must be a pointer to an empty slice of links. A
// valid link type is a struct containing special fields and type annotations.
// Fields of the struct can be:
//
// - Required field of type multicodec.Multicodec with type annotation
//   `ipld:"multihash"`
// - Optional string field containing the link name (last item from the path
//   within the node). it must contain the annotation `ipld:"name"`
// - Extra fields to match extra keys in IPLD Node. The type must be compatible
//   and the type annotation must be `ipld:"key:<extra key>"`
//
// Fields must be publicly accessible.
func ReadLinks(r stream.NodeReader, reslinks interface{}) error {
	rv := reflect.ValueOf(reslinks)

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("Expected slice pointer to Link type: found non pointer type %s", rv.Type().String())
	}

	sv := reflect.Indirect(rv)
	if sv.Kind() != reflect.Slice {
		return fmt.Errorf("Expected slice pointer to Link type: found non slice type %s", sv.Type().String())
	}

	lt := sv.Type().Elem()
	if lt.Kind() != reflect.Struct {
		return fmt.Errorf("Expected slice pointer to Link type: found non struct Link type %s", lt.String())
	}

	var f fields
	f.mapFields(lt, nil)

	if f.link == nil && f.hash == nil {
		return fmt.Errorf("Expected Link type to have a link field (annotated ipld:\"link\") of type string")
	}

	var stack []reflect.Value
	var okStack []bool

	it := stream.Iterate(r, nil)
	defer it.Close()

	for it.Iter() {
		if it.TokenType == stream.TokenNode {
			path := it.StringPath()
			name := ""
			if len(path) > 0 {
				name = path[len(path)-1]
			}
			v := reflect.New(lt)
			if f.name != nil {
				v.Elem().FieldByIndex(f.name).SetString(name)
			}
			if f.spath != nil {
				v.Elem().FieldByIndex(f.spath).Set(reflect.ValueOf(path))
			}
			if f.ipath != nil {
				v.Elem().FieldByIndex(f.ipath).Set(reflect.ValueOf(it.Path))
			}
			stack = append(stack, v)
			okStack = append(okStack, false)

		} else if it.TokenType == stream.TokenKey && it.Value == links.LinkKey {
			it.Iter()
			if f.hash != nil {
				h, err := ReadLinkPath(it.Value)
				if err == nil {
					stack[len(stack)-1].Elem().FieldByIndex(f.hash).Set(reflect.ValueOf(h))
					okStack[len(okStack)-1] = true
				}
			}
			if f.link != nil {
				if slink, ok := it.Value.(string); ok {
					stack[len(stack)-1].Elem().FieldByIndex(f.link).Set(reflect.ValueOf(slink))
					okStack[len(okStack)-1] = true
				}
			}

		} else if key, isKey := it.Value.(string); isKey && it.TokenType == stream.TokenKey {
			if field, has_field := f.attrs[key]; has_field {
				it.Iter()
				setValue(stack[len(stack)-1].Elem().FieldByIndex(field), it)
			}

		} else if it.TokenType == stream.TokenEndNode {
			if okStack[len(okStack)-1] {
				sv.Set(reflect.Append(sv, stack[len(stack)-1].Elem()))
			}
			stack = stack[:len(stack)-1]
			okStack = okStack[:len(okStack)-1]

		}
	}

	return it.LastError
}

func Unmarshal(r stream.NodeReader, dest interface{}) error {
	rv := reflect.ValueOf(dest)

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("Expected pointer to struct: found non pointer type %s", rv.Type().String())
	}

	sv := reflect.Indirect(rv)
	if sv.Kind() != reflect.Struct {
		return fmt.Errorf("Expected pointer to struct: found non struct type %s", sv.Type().String())
	}

	st := sv.Type()
	if st.Kind() != reflect.Struct {
		return fmt.Errorf("Expected pointer to struct: found non struct type %s", st.String())
	}

	it := stream.Iterate(r, nil)
	defer it.Close()

	it.Iter()

	return fillAnyValue(it, sv)
}

func fillAnyValue(it *stream.NodeIterator, v reflect.Value) error {
	if it.TokenType == stream.TokenValue {
		v.Set(reflect.ValueOf(it.Value))
		return nil
	} else if it.TokenType == stream.TokenArray {
		return fillArray(it, v)
	} else if it.TokenType == stream.TokenNode {
		return fillNode(it, v)
	} else {
		return fmt.Errorf("NodeReader: unexpected token %s for value", stream.TokenName(it.TokenType))
	}
}

func fillNode(it *stream.NodeIterator, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("Impossible to assign node %v to non struct %s", it.StringPath(), v.Type().String())
	}

	var f fields
	f.mapFields(v.Type(), nil)

	path := it.StringPath()
	name := ""
	if len(path) > 0 {
		name = path[len(path)-1]
	}

	if f.name != nil {
		v.FieldByIndex(f.name).SetString(name)
	}
	if f.spath != nil {
		v.FieldByIndex(f.spath).Set(reflect.ValueOf(path))
	}
	if f.ipath != nil {
		v.FieldByIndex(f.ipath).Set(reflect.ValueOf(it.Path))
	}

	for {

		if !it.Iter() {
			return fmt.Errorf("unexpected end of NodeReader stream")
		}
		if it.TokenType == stream.TokenEndNode {
			return nil
		} else if it.TokenType != stream.TokenKey {
			return fmt.Errorf("NodeReader: unexpected token %s in Node", stream.TokenName(it.TokenType))
		}

		key, ok := it.ToString()
		if !ok {
			return fmt.Errorf("NodeReader: Cannot convert key %#v to string", it.Value)
		}

		attr, ok := f.attrs[key]
		if !ok && key != links.LinkKey {
			it.Skip()
			continue
		} else if !ok {
			attr = nil
		}

		if !it.Iter() {
			return fmt.Errorf("unexpected end of NodeReader stream")
		}

		if key == links.LinkKey && it.TokenType == stream.TokenValue {
			str, isstr := it.ToString()
			if isstr && f.link != nil {
				v.FieldByIndex(f.link).Set(reflect.ValueOf(str))
			}
			if isstr && f.hash != nil {
				h, err := ReadLinkPath(str)
				if err == nil {
					v.FieldByIndex(f.hash).Set(reflect.ValueOf(h))
				}
			}
		}

		if attr != nil {
			err := fillAnyValue(it, v.FieldByIndex(attr))
			if err != nil {
				return err
			}
		}
	}
}

func fillArray(it *stream.NodeIterator, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	if t.Kind() != reflect.Slice {
		return fmt.Errorf("Impossible to assign array %v to non slice %s", it.StringPath(), t.String())
	}

	for {

		if !it.Iter() {
			return fmt.Errorf("unexpected end of NodeReader stream")
		}
		if it.TokenType == stream.TokenEndArray {
			return nil
		} else if it.TokenType != stream.TokenIndex {
			return fmt.Errorf("NodeReader: unexpected token %s in array", stream.TokenName(it.TokenType))
		}

		if !it.Iter() {
			return fmt.Errorf("unexpected end of NodeReader stream")
		}

		v.Set(reflect.Append(v, reflect.Zero(t.Elem())))

		err := fillAnyValue(it, v.Index(v.Len()-1))
		if err != nil {
			return err
		}
	}
}
