package commit

import (
	"github.com/ipfs/go-ipld/links"
)

// this would serialize to:
//
//   {
//     "@context": "/ipfs/<hash-to-commit-schema>/commit"
//     "parents":   [ {"@link": "<hash1>"}, ... ]
//     "author":    {"@link": "<hash2>"},
//     "committer": {"@link": "<hash3>"},
//     "object":    {"@link": "<hash4>"},
//     "comment": "comment as a string"
//   }
//
type Commit struct {
	Parents   []links.SimpleHashLink //
	Author    links.SimpleHashLink   // link to an Authorship
	Committer links.SimpleHashLink   // link to an Authorship
	Object    links.SimpleHashLink   // what we version ("tree" in git)
	Comment   string                 // describes the commit
}

func (c *Commit) IPLDValidate() bool {
	// check at least one parent exists
	// check Parents have proper type
	// check author exists and has proper type
	// check commiter exists and has proper type
	// check object exists and has proper type
	return true
}
