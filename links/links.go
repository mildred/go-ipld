package links

import (
	"fmt"
	"reflect"

	stream "github.com/ipfs/go-ipld/stream"
	mh "github.com/jbenet/go-multihash"
)

const (
	LinkKey = "@link"
)

// Base type to be inherited to form Link types.
type BaseLink struct {
	Hash       mh.Multihash  `ipld:"multihash"`
	Link       string        `ipld:"link"`
	Name       string        `ipld:"name"`
	Path       []interface{} `ipld:"path"`
	StringPath []string      `ipld:"path"`
}

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
func ReadLinks(r stream.NodeReader, links interface{}) error {
	rv := reflect.ValueOf(links)
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
	f.attrs = make(map[string][]int)
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

		} else if it.TokenType == stream.TokenKey && it.Value == LinkKey {
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
