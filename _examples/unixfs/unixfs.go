package sig

import (
	"os"

	mh "github.com/jbenet/go-multihash"
)

// Dir represents a directory in unixfs. The links are
// dag.Links with one additional property:
//  * unixMode - the full unix mode
//
// this would serialize to:
//
//    {
//      "@context": "/ipfs/<hash-of-schema>/unixdir",
//      "entries": [
//    	  "<filename1>": {
//          "@link": "<hash1>",
//          "unixMode": <mode1>,
//          "size": <size1>
//    	  },
//    	  "<filename2>": {
//          "@link": "<hash2>",
//          "unixMode": <mode2>,
//          "size": <size2>
//        }
//      ]
//    }
//

type EntryLink struct {
	Hash     mh.Multihash `ipld:"multihash"`
	Name     string       `ipld:"name"`
	Size     uint64       `ipld:"key:size"`
	UnixMode uint32       `ipld:"key:unixMode"`
}

type Dir struct {
	Entries map[string]EntryLink `ipld:"key:entries"`
}

func (d *Dir) Entry(e string) (*EntryLink, error) {
	l, ok := d.Entries[e]
	if !ok {
		return nil, os.ErrNotExist
	}
	return &l, nil
}

func (d *Dir) Mode(e string) (os.FileMode, error) {
	l, ok := d.Entries[e]
	if !ok {
		return 0, os.ErrNotExist
	}

	return (os.FileMode)(l.UnixMode), nil
}
