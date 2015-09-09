package ipld

import (
	"errors"
	"strconv"
	"strings"
)

const pathSep = "/"

// pathEscape escapes the / and \ characters of path components
func pathEscape(s string) string {
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "/", "\\/", -1)
	return s
}

// JoinPath will join path elements together and escape the /
// character of each path component.
func JoinPath(path []string) string {
	p := []string{}
	for _, s := range path {
		p = append(p, pathEscape(s))
	}
	return strings.Join(p, pathSep)
}

// SplitPath will split a path on / characters and return a string
// array. The / character can be escaped with \ in which case it
// doesn't count as a separator. Empty elements (leading, trailing
// and duplicate separators) are ignored.
func SplitPath(path string) []string {
	comps := []string{}
	comp  := []byte{}
	esc   := false
	for _, c := range []byte(path) {
		if !esc && c == '/' {
			if len(comp) > 0 {
				comps = append(comps, string(comp))
			}
			comp = []byte{}
		} else if c == '\\' {
			esc = true
		} else {
			esc = false
			comp = append(comp, c)
		}
	}
	return comps
}

// SkipNode is a special value used with Walk and WalkFunc.
// If a WalkFunc returns SkipNode, the walk skips the curr
// node and its children. It behaves like file/filepath.SkipDir
var SkipNode = errors.New("skip node from Walk")

// WalkFunc is the type of the function called for each node
// visited by Walk. The root argument is the node from which
// the Walk began. The curr argument is the currently visited
// node. The path argument is the traversal path, from root
// to curr.
//
// If there was a problem walking to curr, the err argument
// will describe the problem and the function can decide
// how to handle the error (and Walk _will not_ descend into
// any of the children of curr).
//
// WalkFunc may return a node, in which case the returned node
// will be used for further traversal instead of the curr
// node.
//
// WalkFunc may return an error. If the error is the special
// SkipNode error, the children of curr are skipped. All other
// errors halt processing early. In this respect, it behaves
// just like file/filepath.WalkFunc
type WalkFunc func(root, curr Node, path []string, err error) (Node, error)

// Walk traverses the given root node and all its children, calling
// WalkFunc with every Node visited, including root. All errors
// that arise while visiting nodes are passed to given WalkFunc.
// The order in which children are visited is not deterministic.
// Walk traverses sequences as well, which is to mean the nodes
// below will be visted as "foo/0", "foo/1", and "foo/3":
//
//   { "foo": [
//     {"a":"aaa"}, // visited as foo/0
//     {"b":"bbb"}, // visited as foo/1
//     {"c":"ccc"}, // visited as foo/2
//   ]}
//
// Walk returns a node constructed from the transformed result of the
// walk function.
//
// Note Walk is purely local and does not traverse Links. For a
// version of Walk that does traverse links, see the ipld/traverse
// package.
func Walk(root Node, walkFn WalkFunc) (Node, error) {
	n, err := walk(root, root, nil, walkFn)
	if node, ok := n.(Node); ok {
		return node, err
	} else {
		return nil, err
	}
}

// WalkFrom is just like Walk, but starts the Walk at given startFrom
// sub-node. It is the equivalent of a regular Walk call which skips
// all nodes which do not have startFrom as a prefix.
func WalkFrom(root Node, startFrom []string, walkFn WalkFunc) (interface{}, error) {
	start := GetPath(root, startFrom)
	if start == nil {
		return nil, errors.New("no descendant at " + JoinPath(startFrom))
	}
	return walk(root, start, startFrom, walkFn)
}

// walk is used to implement Walk.
func walk(root Node, curr interface{}, npath []string, walkFunc WalkFunc) (interface{}, error) {

	if nc, ok := curr.(Node); ok { // it's a node!
		// first, call user's WalkFunc.
		newnode, err := walkFunc(root, nc, npath, nil)
		res := Node{}
		if err == SkipNode {
			return newnode, nil // ok, let's skip this one.
		} else if err != nil {
			return nil, err // something bad happened, return early.
		} else if newnode != nil {
			nc = newnode
		}

		// then recurse.
		for k, v := range nc {
			n, err := walk(root, v, append(npath, k), walkFunc)
			if err != nil {
				return nil, err
			} else if n != nil {
				res[k] = n
			}
		}

		return res, nil

	} else if sc, ok := curr.([]interface{}); ok { // it's a slice!
		res := []interface{}{}
		for i, v := range sc {
			k := strconv.Itoa(i)
			n, err := walk(root, v, append(npath, k), walkFunc)
			if err != nil {
				return nil, err
			} else if n != nil {
				res = append(res, n)
			}
		}
		return res, nil

	} else { // it's just data.
		// ignore it.
	}
	return curr, nil
}

// GetPathCmp gets a descendant of root, at npath.
func GetPath(root interface{}, npath []string) interface{} {
	if len(npath) == 0 {
		return root // we're done.
	}
	if root == nil {
		return nil // nowhere to go
	}

	k := npath[0]
	if vn, ok := root.(Node); ok {
		// if node, recurse
		return GetPath(vn[k], npath[1:])

	} else if vs, ok := root.([]interface{}); ok {
		// if slice, use key as an int offset
		i, err := strconv.Atoi(k)
		if err != nil {
			return nil
		}
		if i < 0 || i >= len(vs) { // nothing at such offset
			return nil
		}

		return GetPath(vs[i], npath[1:])
	}

	return nil // cannot keep walking...
}
