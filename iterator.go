package ipld

import (
	"math/big"
)

type NodeIterator struct {
	ReadItem
	LastError      error
	nextError      bool
	items          chan *ReadItem
	feedback       chan error
	feedbackClosed bool
}

type ReadItem struct {
	Path      []interface{}
	TokenType ReaderToken
	Value     interface{}
}

func (s *ReadItem) StringPath() []string {
	return ToStringPath(s.Path)
}

func (s *ReadItem) ToInt() (res int64, ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	switch s.Value.(type) {
	case int:
		res = int64(s.Value.(int))
	case int64:
		res = s.Value.(int64)
	case uint64:
		res = int64(s.Value.(uint64))
	case *big.Int:
		i := s.Value.(*big.Int)
		if i.BitLen() > 63 {
			ok = false
		} else {
			res = i.Int64()
		}
	default:
		ok = false
	}
	return res, ok
}

func (s *ReadItem) ToUint() (res uint64, ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	switch s.Value.(type) {
	case int:
		res = uint64(s.Value.(int))
	case int64:
		res = uint64(s.Value.(int64))
	case uint64:
		res = s.Value.(uint64)
	case *big.Int:
		i := s.Value.(*big.Int)
		if i.BitLen() > 64 || i.Sign() < 0 {
			ok = false
		} else {
			res = i.Uint64()
		}
	default:
		ok = false
	}
	return res, ok
}

func (s *ReadItem) ToFloat() (res float64, ok bool) {
	ok = true
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	switch s.Value.(type) {
	case int:
		res = float64(s.Value.(int))
	case int64:
		res = float64(s.Value.(int64))
	case uint64:
		res = float64(s.Value.(uint64))
	case float32:
		res = float64(s.Value.(float32))
	case float64:
		res = s.Value.(float64)
	case *big.Int:
		i := s.Value.(*big.Int)
		if i.BitLen() > 64 {
			ok = false
		} else if i.Sign() < 0 {
			res = -float64(big.NewInt(0).Abs(i).Uint64())
		} else {
			res = float64(i.Uint64())
		}
	default:
		ok = false
	}
	return res, ok
}

// Read from a NodeReader using a channel of ReadItem.
func Iterate(r NodeReader, res_error *error) *NodeIterator {
	it := make(chan *ReadItem)
	ce := make(chan error)
	res := &NodeIterator{ReadItem{}, nil, true, it, ce, false}

	go func() {
		err := r.Read(func(path []interface{}, tokenType ReaderToken, value interface{}) error {
			item := &ReadItem{path, tokenType, value}
			it <- item
			return <-ce
		})
		close(it)
		if err != nil && res_error != nil {
			res.LastError = err
		}
	}()

	return res
}

func (s *NodeIterator) Iter() bool {
	if !s.nextError {
		s.feedback <- nil
	}

	item := <-s.items
	if item == nil {
		if !s.feedbackClosed {
			close(s.feedback)
			s.feedbackClosed = true
		}
		return false
	}

	s.ReadItem = *item
	return true
}

func (s *NodeIterator) Abort() {
	s.StopError(NodeReadAbort)
}

func (s *NodeIterator) Skip() {
	s.StopError(NodeReadSkip)
}

func (s *NodeIterator) StopError(e error) {
	if s.nextError {
		panic("Already stopped")
	}

	s.feedback <- e
	s.nextError = true
}

func (s *NodeIterator) Close() error {
	if !s.nextError {
		s.Abort()
	}
	for s.Iter() {
		s.Abort()
	}
	return s.LastError
}
