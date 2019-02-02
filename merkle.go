package merklelib

import (
	"math"
	"math/bits"
)

var (
	ErrInvalidHasherType = error.New("Invalid hasher object.")
	ErrInvalidAuditProof = error.New("Invalid audit proof object.")
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

type AuditProofNode struct {
	Hash []byte
	pos  NodeType
}

type AuditProof struct {
	Nodes []AuditProofNode
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
	return &Tree{hasher, nil, make(map[string]*MerkleNode)}
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
func (tree *Tree) Extend(data ...[]interface{}) {
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
				newNode := mergeNodes(tree.hasher, nodes[i-1], nodes[i])
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
	var parent, sibling *MerkleNode
	for node != tree.root {
		parent = node.parent
		sibiling = node.sibiling()
		parent.hashVal = concatHashes(tree.hasher, node, sibiling)
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
	tree.mapping[string(newHash)] = node

	if last == tree.root {
		tree.root = mergeNodes(tree.hasher, tree.root, last)
		return
	}

	sibiling := last.sibling()
	connector := last.parent

	if sibling == sentinel {
		node.parent = connector
		connector.right = node
		tree.rehash(node)
		return
	}

	node.right = sentinel
	for connector != tree.root {
		node = mergeNodes(tree.hasher, node, sentinel)
		sibiling = connector.sibiling

		if sibiling == sentinel {
			connector.parent.right = node
			node.parent = connector.parent
			tree.rehash(node)
			return
		}

		connector = connector.parent
	}

	node = mergeNodes(tree.hasher, node, sentinel)
	tree.root = mergeNodes(tree.hasher, connector, node)
}

func (tree *Tree) getLeaf(leaf interface{}) *MerkleNode {
	key := toBytes(old)
	leaf := tree.mapping.Get(string(key))
	// try hashing
	if leaf == nil {
		key = tree.hasher.HashLeaf(key)
		leaf = tree.mapping.Get(string(key))
	}
	return leaf
}

// updating items
func (tree *Tree) Update(old, new interface{}) {
	leaf := tree.getLeaf(old)
	if leaf != nil {
		tree.mapping.Delete(string(leaf.hashVal))
		hashVal := tree.hasher.HashLeaf(new)
		leaf.hashVal = hashVal
		tree.mapping.Add(string(hashVal), leaf)
		tree.rehash(leaf)
	}
	// TODO: throw an error
}

// get an audit proof that an item / leaf is in the tree
func (tree *Tree) GetProof(leaf interface{}) *AuditProof {
	leaf := tree.getLeaf(leaf)
	if leaf != nil {
		paths := make([]AuditProofNode)
		for leaf != tree.root {
			sibiling := leaf.sibling()
			if sibling != sentinel {
				node := AuditProofNode(sibling.hashVal, sibiling.Position())
				paths = append(paths, node)
			}
			leaf = leaf.parent
		}
		return &AuditProof{paths}
	}
}

func (tree *Tree) Clear() {
	tree.root = nil
	tree.mapping.Clear()
}

// TODO: move to the utils file?
func toHasher(item interface{}) (*Hasher, error) {
	switch value := item.(type) {
	case Hasher:
		return value, nil
	case HashFunc:
		return &Hasher{value}, nil
	default:
		return nil, ErrInvalidHasherType
	}
}

func (tree *Tree) Size() uint {
	return tree.mapping.size()
}

func (tree *Tree) VerifyTreeConsistency(oldRootHash []byte, oldTreeSize uint) bool {
	treeSize := tree.Size()
	if treeSize < oldTreeSize {
		return false
	}

	rootHash := tree.root.hashVal

	if treeSize == oldTreeSize {
		return rootHash == oldRootHash
	}

	leaves := tree.mapping.Values()
	index, paths := 0, make([]*MerkleNode)

	for oldTreeSize > 0 {
		level := math.Exp2(float64(bits.Len(oldTreeSize)))
		node := leaves[index].climbNode(uint(math.log2(level)))
		if node == nil {
			return false
		}
		paths = append(paths, node)
		index += uint(level)
		oldTreeSize -= uint(level)
	}

	return paths[0].hashVal == oldRootHash
}

func VerifyLeafInclusion(leaf interface{}, proof *AuditProof, hasher interface{}, rootHash []byte) (bool, error) {
	if len(proof.Nodes) < 1 {
		return false, ErrInvalidAuditProof
	}
	nodes := proof.Nodes
	// if hasher is a hash function, convert it to hasher
	hasher, err := toHasher(hasher)
	if err != nil {
		return false, err
	}
	// try without hashing
	leaf, err := toBytes(leaf)
	if err != nil {
		return false, err
	}

	newRoot := leaf
	for _, right := range nodes[1:] {
		// TODO: concat hashes
	}

	return newRoot == rootHash
}
