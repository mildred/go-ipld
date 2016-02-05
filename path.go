package ipld

import (
	"fmt"
)

// Convert a []interface{} path to a []string path
// This should only convert array indices to strings, and such are not ambiguous
// because array indices and object keys cannot be mixed for the same prefix.
func ToStringPath(path []interface{}) []string {
	res := make([]string, len(path))
	for _, e := range path {
		if str, ok := e.(string); ok {
			res = append(res, str)
		} else if i, ok := e.(int); ok {
			res = append(res, fmt.Sprintf("%d", i))
		} else if i, ok := e.(uint64); ok {
			res = append(res, fmt.Sprintf("%d", i))
		} else {
			res = append(res, fmt.Sprintf("%v", e))
		}
	}
	return res
}
