package test

import (
	"testing"

	reader "github.com/ipfs/go-ipld"
	"github.com/mildred/assrt"
)

type Callback struct {
	Path      []interface{}
	TokenType reader.ReaderToken
	Value     interface{}
}

func CheckReader(test *testing.T, node reader.NodeReader, callbacks []Callback) {
	assert := assrt.NewAssert(test)
	var i int = 0
	err := node.Read(func(path []interface{}, tokenType reader.ReaderToken, value interface{}) error {
		if i >= len(callbacks) {
			assert.Logf("Callback %d: not described in test", i)
			assert.Logf("Should be: {%#v, %v, %#v}", path, reader.TokenName(tokenType), value)
			assert.Fail()
		} else {
			cb := callbacks[i]
			etk := reader.TokenName(tokenType)
			atk := reader.TokenName(cb.TokenType)
			actual := Callback{path, tokenType, value}
			assert.Logf("Callback %3d: %s %#v", i, etk, cb)
			assert.Logf("         got: %s %#v", atk, actual)
			assert.Equal(cb, actual)
		}
		i++
		return nil
	})
	assert.Equal(len(callbacks), i, "Number of callbacks incorrect")
	assert.Nil(err)
}
