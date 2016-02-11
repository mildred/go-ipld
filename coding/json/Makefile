stream-go1.4.go: stream.go Makefile
	sed -r -e '/\+build/ { s/go/!go/ }' -e 's:encoding/json:github.com/ipfs/go-ipld/Godeps/_workspace/src/github.com/golang/go/src/encoding/json:' $< >$@
