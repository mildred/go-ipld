package links

import (
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

type SimpleLink struct {
	Link string `ipld:"link"`
}

type SimpleHashLink struct {
	Hash mh.Multihash `ipld:"multihash"`
}
