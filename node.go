package ipld

import (
	"errors"
	"math/big"
)

var ErrEnd = errors.New("IpldIterateEnd")

type Node interface{}

type NodeObject interface {
	Next() (key string, value Node, err error)
}

type NodeObjectMemory interface {
	NodeObject
	Reset()
	Length() uint64
	Get(key string) Node
}

type NodeLink interface {
	LinkPath() string // should be multiaddr ???
}

type NodeObjectLink interface {
	NodeObject
	NodeLink
}

type NodeObjectLinkMemory interface {
	NodeObjectMemory
	NodeLink
}

type NodeArray interface {
	Next() (value Node, err error)
}

type NodeArrayMemory interface {
	NodeArray
	Reset()
	Length() uint64
	Get(idx uint64) Node
}

func ToObject(n Node) NodeObject {
	return n.(NodeObject)
}

func ToObjectMemory(n Node) NodeObjectMemory {
	return n.(NodeObjectMemory)
}

func ToObjectLink(n Node) NodeObjectLink {
	return n.(NodeObjectLink)
}

func ToObjectLinkMemory(n Node) NodeObjectLinkMemory {
	return n.(NodeObjectLinkMemory)
}

func ToArray(n Node) NodeArray {
	return n.(NodeArray)
}

func ToArrayMemory(n Node) NodeArrayMemory {
	return n.(NodeArrayMemory)
}

func ToBytes(n Node) []byte {
	var ok bool
	var res []byte
	switch n.(type) {
	case string:
		res = []byte(n.(string))
		ok = true
	case []byte:
		res = n.([]byte)
		ok = true
	default:
		ok = false
	}
	if res == nil {
		res = []byte{}
	}
	if ok {
		return res
	} else {
		return nil
	}
}

func ToString(n Node) *string {
	var ok bool
	var res string
	switch n.(type) {
	case string:
		res = n.(string)
		ok = true
	case []byte:
		res = string(n.([]byte))
		ok = true
	default:
		ok = false
	}
	if ok {
		return &res
	} else {
		return nil
	}
}

func ToUint(n Node) *uint64 {
	var res uint64
	var ok bool
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	switch n.(type) {
	case int:
		res = uint64(n.(int))
		ok = true
	case int64:
		res = uint64(n.(int64))
		ok = true
	case uint64:
		res = n.(uint64)
		ok = true
	case *big.Int:
		i := n.(*big.Int)
		if i.BitLen() <= 64 && i.Sign() >= 0 {
			res = i.Uint64()
			ok = true
		}
	default:
	}
	if ok {
		return &res
	} else {
		return nil
	}
}

func ToInt(n Node) *int64 {
	var res int64
	var ok bool
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	switch n.(type) {
	case int:
		res = int64(n.(int))
		ok = true
	case int64:
		res = n.(int64)
		ok = true
	case uint64:
		res = int64(n.(uint64))
		ok = true
	case *big.Int:
		i := n.(*big.Int)
		if i.BitLen() <= 63 {
			res = i.Int64()
			ok = true
		}
	default:
	}
	if ok {
		return &res
	} else {
		return nil
	}
}

func ToFloat(n Node) *float64 {
	var ok bool
	var res float64
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	switch n.(type) {
	case int:
		res = float64(n.(int))
		ok = true
	case int64:
		res = float64(n.(int64))
		ok = true
	case uint64:
		res = float64(n.(uint64))
		ok = true
	case float32:
		res = float64(n.(float32))
		ok = true
	case float64:
		res = n.(float64)
		ok = true
	case *big.Int:
		i := n.(*big.Int)
		if i.BitLen() <= 64 {
			if i.Sign() < 0 {
				res = -float64(big.NewInt(0).Abs(i).Uint64())
				ok = true
			} else {
				res = float64(i.Uint64())
				ok = true
			}
		}
	default:
	}
	if ok {
		return &res
	} else {
		return nil
	}
}
