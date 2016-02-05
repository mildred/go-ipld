package ipld

import (
	"fmt"
	"reflect"
	"strings"
)

// A NodeReader that only read elements from a path prefix
type NodeReaderAt struct {
	parent NodeReader
	prefix []string
}

func (n *NodeReaderAt) Read(f ReadFun) error {
	found := false
	return n.parent.Read(func(path []interface{}, tokenType ReaderToken, value interface{}) error {
		if isPathPrefix(n.prefix, ToStringPath(path)) {
			found = true
			return f(path[len(n.prefix):], tokenType, value)
		} else if found {
			// Finished, stop early
			return NodeReadAbort
		} else {
			// Not yet here, continue
			return nil
		}
	})
}

// Return true if path has prefix
func isPathPrefix(prefix, path []string) bool {
	return len(prefix) <= len(path) && reflect.DeepEqual(prefix, path[0:len(prefix)])
}

// Return an instance of NodeReaderAt that read root only starting at the given
// prefix.
func GetReaderAt(root NodeReader, path string) *NodeReaderAt {
	return &NodeReaderAt{root, strings.Split(path, "/")}
}
