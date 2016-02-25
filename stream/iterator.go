package stream

import (
	"fmt"
	"math/big"
)

type ReadItem struct {
	Path      []interface{}
	TokenType ReaderToken
	Value     interface{}
}

// Convert a []interface{} path to a []string path
// This should only convert array indices to strings, and such are not ambiguous
// because array indices and object keys cannot be mixed for the same prefix.
func ToStringPath(path []interface{}) []string {
	res := make([]string, len(path))
	for i, e := range path {
		if str, ok := e.(string); ok {
			res[i] = str
		} else if ival, ok := e.(int); ok {
			res[i] = fmt.Sprintf("%d", ival)
		} else if uival, ok := e.(uint64); ok {
			res[i] = fmt.Sprintf("%d", uival)
		} else {
			res[i] = fmt.Sprintf("%v", e)
		}
	}
	return res
}

func (s *ReadItem) StringPath() []string {
	return ToStringPath(s.Path)
}

func (s *ReadItem) ToString() (string, bool) {
	switch s.Value.(type) {
	case string:
		return s.Value.(string), true
	case []byte:
		return string(s.Value.([]byte)), true
	default:
		return "", false
	}
}

func (s *ReadItem) ToBytes() ([]byte, bool) {
	switch s.Value.(type) {
	case string:
		return []byte(s.Value.(string)), true
	case []byte:
		return s.Value.([]byte), true
	default:
		return nil, false
	}
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

type NodeIterator struct {
	*ReadItem
	LastError      error
	needFeedback   bool
	items          chan *ReadItem
	feedback       chan error
	feedbackClosed bool
}

// Read from a NodeReader using a channel of ReadItem.
func Iterate(r NodeReader, res_error *error) *NodeIterator {
	it := make(chan *ReadItem)
	feedback := make(chan error)
	res := &NodeIterator{&ReadItem{}, nil, false, it, feedback, false}

	go func() {
		err := r.Read(func(path []interface{}, tokenType ReaderToken, value interface{}) error {
			item := &ReadItem{path, tokenType, value}
			it <- item
			return <-feedback
		})
		close(it)
		if err != nil && res_error != nil {
			res.LastError = err
		}
	}()

	return res
}

func (s *NodeIterator) Iter() bool {
	if s.needFeedback {
		s.feedback <- nil
	}

	item := <-s.items
	if item == nil {
		s.needFeedback = false
		if !s.feedbackClosed {
			close(s.feedback)
			s.feedbackClosed = true
		}
		return false
	}

	s.needFeedback = true
	s.ReadItem = item
	return true
}

func (s *NodeIterator) Valid() bool {
	return s.needFeedback
}

func (s *NodeIterator) Abort() {
	s.StopError(NodeReadAbort)
}

func (s *NodeIterator) Skip() {
	s.StopError(NodeReadSkip)
}

func (s *NodeIterator) StopError(e error) {
	if !s.needFeedback {
		panic("Already stopped")
	}

	s.feedback <- e
	s.needFeedback = false
}

func (s *NodeIterator) Close() error {
	if s.needFeedback {
		s.Abort()
	}
	for s.Iter() {
		s.Abort()
	}
	return s.LastError
}
