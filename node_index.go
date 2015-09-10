package ipld

import (
	"strings"
)

func isContainerIndex(n Node) bool {
	return n["@container"] == "@index"
}

const defaultIndexName string = "@index"

func containerIndexName(n Node, defaultval string) string {
	var index_name string = defaultval

	index_val, ok := n["@index"]
	if str, is_string := index_val.(string); ok && is_string {
		index_name = str
	}

	return index_name
}

func copyNode(n Node) Node {
	var res Node = Node{}
	for k, v := range n {
		res[k] = v
	}
	return res
}

func ParseNodeIndex(n Node) (attrs, directives, index Node, escapedIndex Node) {
	attrs = Node{}
	directives = Node{}
	index = Node{}
	escapedIndex = Node{}

	if real_attrs, ok := n["@attrs"]; ok {
		if attrs_node, ok := real_attrs.(Node); ok {
			attrs = copyNode(attrs_node)
		}
	}

	index_container := isContainerIndex(n)

	for key, val := range n {
		if key == "@attrs" {
			continue
		} else if key[0] == '@' {
			if key == "@index" {
				attrs[key] = val
			} else {
				directives[key] = val
			}
		} else {
			if index_container {
				escapedIndex[key] = val
				index[unescapeFromDirective(key)] = val
			} else {
				attrs[unescapeFromDirective(key)] = val
			}
		}
	}

	return
}

func unescapeFromDirective(key string) string {
	key = strings.Replace(key, "\\@", "@", -1)
	key = strings.Replace(key, "\\\\", "\\", -1)
	return key
}

