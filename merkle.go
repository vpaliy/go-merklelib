package merklelib

import (
	"bytes"
	"errors"
	"math"
	"math/bits"
)

var (
	ErrInvalidHasherType = errors.New("Invalid hasher object.")
	ErrInvalidAuditProof = errors.New("Invalid audit proof object.")
)

// an alias for a hash function
type HashFunc func([]byte) []byte

type Hasher struct {
	hashFunc HashFunc
}

type Tree struct {
	hasher  Hasher
	root    *MerkleNode
	mapping OrderedMap
}

type AuditProof struct {
	Nodes []*AuditProofNode
}

func (hasher *Hasher) HashLeaf(leaf interface{}) []byte {
	buffer := toBytes(leaf)
	return hasher.hashFunc(append(buffer, 0x00))
}

// hash children together by adding 0x01 byte
func (hasher *Hasher) HashChildren(left, right interface{}) []byte {
	buffer := append(toBytes(left), toBytes(right)...)
	return hasher.hashFunc(append(buffer, 0x01))
}

// creates a new Merkle tree with the provided hasher
func New(hashFunc HashFunc) *Tree {
	hasher := Hasher{hashFunc}
	dict := OrderedMap{nil, nil, make(map[string]*mapnode)}
	return &Tree{hasher, nil, dict}
}

// return the merkle hash root
func (tree *Tree) Root() []byte {
	if tree.root != nil {
		return tree.root.hashVal
	}
	return nil
}

// retrieve the values from the underlying map object
func (tree *Tree) Leaves() [][]byte {
	if tree.root != nil {
		nodes := tree.mapping.Values()
		leaves := make([][]byte, len(nodes))
		for _, node := range nodes {
			leaves = append(leaves, node.hashVal)
		}
		return leaves
	}
	return nil
}

// append multiple nodes to the tree
func (tree *Tree) Extend(data ...interface{}) {
	if tree.root == nil {
		nodes := make([]*MerkleNode, len(data))
		mapping := tree.mapping
		for index, item := range data {
			hashVal := tree.hasher.HashLeaf(item)
			node := newNode(hashVal)
			mapping.Add(string(hashVal), node)
			nodes[index] = node
		}

		for len(nodes) > 1 {
			if len(nodes)%2 == 0 {
				nodes = append(nodes, sentinel)
			}
			// compute the next level of nodes
			var temp []*MerkleNode
			for i := 1; i < len(nodes); i += 2 {
				newNode := mergeNodes(&tree.hasher, nodes[i-1], nodes[i])
				temp = append(temp, newNode)
			}
			nodes = temp
		}

		tree.root = nodes[0]
	} else {
		// otherwise just append them
		for item := range data {
			tree.Append(item)
		}
	}
}

// rehash some part of the tree starting from a specific node
func (tree *Tree) rehash(node *MerkleNode) {
	for node != tree.root {
		parent := node.parent
		sibling := node.sibling()
		parent.hashVal = concat(&tree.hasher, node, sibling)
		node = parent
	}
}

// append an additional node
func (tree *Tree) Append(item interface{}) {
	if tree.root == nil {
		node := newNode(tree.hasher.HashLeaf(item))
		tree.root = node
		tree.mapping.Add(string(node.hashVal), node)
		return
	}

	last := tree.mapping.Last()
	newHash := tree.hasher.HashLeaf(item)
	node := newNode(newHash)
	tree.mapping.Add(string(newHash), node)

	if last == tree.root {
		tree.root = mergeNodes(&tree.hasher, tree.root, last)
		return
	}

	sibling := last.sibling()
	connector := last.parent

	if sibling == sentinel {
		node.parent = connector
		connector.right = node
		tree.rehash(node)
		return
	}

	node.right = sentinel
	for connector != tree.root {
		node = mergeNodes(&tree.hasher, node, sentinel)
		sibling = connector.sibling()

		if sibling == sentinel {
			connector.parent.right = node
			node.parent = connector.parent
			tree.rehash(node)
			return
		}

		connector = connector.parent
	}

	node = mergeNodes(&tree.hasher, node, sentinel)
	tree.root = mergeNodes(&tree.hasher, connector, node)
}

func (tree *Tree) getLeaf(leaf interface{}) *MerkleNode {
	key := toBytes(leaf)
	node := tree.mapping.Get(string(key))
	// try hashing
	if node == nil {
		key = tree.hasher.HashLeaf(key)
		node = tree.mapping.Get(string(key))
	}
	return node
}

// updating items
func (tree *Tree) Update(old, new interface{}) error {
	leaf := tree.getLeaf(old)
	if leaf != nil {
		tree.mapping.Remove(string(leaf.hashVal))
		hashVal := tree.hasher.HashLeaf(new)
		leaf.hashVal = hashVal
		tree.mapping.Add(string(hashVal), leaf)
		tree.rehash(leaf)
		return nil
	}
	return errors.New("Provided leaf value does not exist in the tree.")
}

// get an audit proof that an item / leaf is in the tree
func (tree *Tree) GetProof(item interface{}) *AuditProof {
	leaf := tree.getLeaf(item)
	if leaf != nil {
		var paths []*AuditProofNode
		for leaf != tree.root {
			sibling := leaf.sibling()
			if sibling != sentinel {
				node := sibling.createAuditProofNode()
				paths = append(paths, node)
			}
			leaf = leaf.parent
		}
		return &AuditProof{paths}
	}
	// TODO: consider returning an error
	return nil
}

func (tree *Tree) Clear() {
	tree.root = nil
	tree.mapping.Clear()
}

// TODO: move to the utils file?
func toHasher(item interface{}) (*Hasher, error) {
	switch value := item.(type) {
	case Hasher:
		return &value, nil
	case *Hasher:
		return value, nil
	case HashFunc:
		return &Hasher{value}, nil
	default:
		return nil, ErrInvalidHasherType
	}
}

func (tree *Tree) Size() int {
	return tree.mapping.Len()
}

func (tree *Tree) VerifyTreeConsistency(oldRootHash []byte, oldTreeSize int) bool {
	treeSize := tree.Size()
	if treeSize < oldTreeSize {
		return false
	}

	rootHash := tree.root.hashVal

	if treeSize == oldTreeSize {
		return bytes.Equal(rootHash, oldRootHash)
	}

	leaves := tree.mapping.Values()
	index := 0
	var paths []*MerkleNode

	for oldTreeSize > 0 {
		level := math.Exp2(float64(bits.Len(uint(oldTreeSize))))
		node := leaves[index].climbTo(int(math.Log2(level)))
		if node == nil {
			return false
		}
		paths = append(paths, node)
		index += int(level)
		oldTreeSize -= int(level)
	}

	var newHashRoot []byte

	if len(paths) > 1 {
		reversedPaths := reverse(paths)
		newHashRoot = reduceToBytes(&tree.hasher, reversedPaths)
	} else {
		newHashRoot = paths[0].hashVal
	}

	return bytes.Equal(newHashRoot, oldRootHash)
}

func VerifyLeafInclusion(item interface{}, proof *AuditProof, hasherObj interface{}, rootHash []byte) (bool, error) {
	if len(proof.Nodes) < 1 {
		return false, ErrInvalidAuditProof
	}
	nodes := proof.Nodes
	// if hasher is a hash function, convert it to hasher
	hasher, err := toHasher(hasherObj)
	if err != nil {
		return false, err
	}
	// try without hashing
	leaf := toBytes(item)
	if err != nil {
		return false, err
	}

	newHashRoot := reduceToBytes(hasher, nodes)
	if !bytes.Equal(newHashRoot, rootHash) {
		leaf = hasher.HashLeaf(leaf)
		newHashRoot = reduceToBytes(hasher, nodes)
	}
	return bytes.Equal(newHashRoot, rootHash), nil
}

func reduceToBytes(hasher *Hasher, data ...interface{}) []byte {
	if len(data) < 2 {
		// TODO: return an error instead
		return nil
	}

	result := concat(hasher, data[0], data[1])

	for _, item := range data[2:] {
		result = concat(hasher, result, item)
	}
	return result
}
