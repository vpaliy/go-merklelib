package merklelib

import (
  "fmt"
  "bytes"
  "crypto"
  "errors"
)

// an alias for a hash function
type HashFunc func(interface{}) []byte

// Building block
type Node struct {
  hashval []byte
  left, right, parent *Node
}

type Hasher struct {
  hash HashFunc
}

type Tree struct {
  hasher Hasher
  root *Node
  leaves map[[]bytes]*Node
}

func New(hashfunc HashFunc) *Tree {
  hasher := Hasher { hashfunc }
  return &Tree { hasher }
}
