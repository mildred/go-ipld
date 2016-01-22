package ipld

import (
	"fmt"
	"reflect"
	"strings"
)

// Convert a []interface{} path to a []string path
// This should only convert array indices to strings, and such are not ambiguous
// because array indices and object keys cannot be mixed for the same prefix.
func ToStringPath(anyPath []interface{}) []string {
	var res []string
	for _, e := range anyPath {
		if str, ok := e.(string); ok {
			res = append(res, str)
		} else if i, ok := e.(int); ok {
			res = append(res, fmt.Sprintf("%d", i))
		} else {
			res = append(res, fmt.Sprintf("%v", e))
		}
	}
	return res
}

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
