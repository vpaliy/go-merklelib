package merklelib

import (
  "encoding/gob"
  "bytes"
  "encoding/hex"
)

var sentinel = new(Node)

func toBytes(item interface{}) []byte {
  switch value := item.(type) {
  case string:
    return []byte(value)
  case []byte:
    return value
  }
  var buf bytes.Buffer
  enc := gob.NewEncoder(&buf)
  err := enc.Encode(item)
  // TODO: check this error here
  if err != nil {
    return nil
  }
  return buf.Bytes()
}

func toHex(src []byte) []byte {
  dst := make([]byte, hex.EncodedLen(len(src)))
  hex.Encode(dst, src)
  return dst
}

type NodeType int16

const (
  LEFT  NodeType = 0
  RIGHT NodeType = 1
  LEAF  NodeType = 2
)

// an alias for a hash function
type HashFunc func([] byte) []byte

// Building block
type Node struct {
  hashVal []byte
  pos NodeType
  left, right, parent *Node
}

type Hasher struct {
  hashfunc HashFunc
}

type Tree struct {
  hasher Hasher
  root *Node
  mapping map[string]*Node
}

type AuditProofNode struct {
  Hash []byte
  pos NodeType
}

type AuditProof struct {
  Nodes []AuditProofNode
}

func newNode(hashVal []byte) *Node {
  return &Node { hashVal, LEAF, nil, nil, nil }
}

func (hasher *Hasher) HashLeaf(leaf interface{}) []byte {
  buffer := toBytes(leaf)
  return hasher.hashfunc(append(buffer, 0x00))
}

// hash children together by adding 0x01 byte
func (hasher *Hasher) HashChildren(left, right interface{}) []byte {
  buffer := append(toBytes(left), toBytes(right)...)
  return hasher.hashfunc(append(buffer, 0x01))
}

// creates a new Merkle tree with the provided hasher
func New(hashfunc HashFunc) *Tree {
  hasher := Hasher { hashfunc }
  return &Tree { hasher, nil, make(map[string]*Node) }
}

// return the merkle hash root
func (tree *Tree) Root() []byte {
  return nil
}

// retrieve the values from the underlying map object
func (tree *Tree) Leaves() [][]byte {
  return nil
}

// append multiple nodes to the tree
func (tree *Tree) Extend(data...[]interface{}) {
  if tree.root == nil {
    nodes := make([]*Node, len(data))
    for index, item := range data {
      hashVal := tree.hasher.HashLeaf(item)
      node := newNode(hashVal)
      tree.mapping[string(hashVal)] = node
      nodes[index] = node
    }

    for len(nodes) > 1 {
      if len(nodes) % 2 == 0 {
        nodes = append(nodes, sentinel)
      }
      // compute the next level of nodes
      var temp []*Node
      for i := 0; i < len(nodes); i += 2 {
        temp = append(temp, nodes[i])
      }
      nodes = temp
    }

    tree.root = nodes[0]
  }
}

// append an additional node
func (tree *Tree) Append(item interface{}) {

}

// updating only real byte, nothing else there
func (tree *Tree) Update(old, new []byte) {

}

// remove a single node from the tree
func (tree *Tree) Remove(item []byte) {

}

// get an audit proof that an item / leaf is in the tree
func (tree *Tree) GetProof(leaf interface{}) {

}

func (tree *Tree) Clear() {

}
