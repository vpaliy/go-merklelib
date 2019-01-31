package merklelib

// an alias for a hash function
type HashFunc func([]byte) []byte

type Hasher struct {
	hashfunc HashFunc
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
	return hasher.hashfunc(append(buffer, 0x00))
}

// hash children together by adding 0x01 byte
func (hasher *Hasher) HashChildren(left, right interface{}) []byte {
	buffer := append(toBytes(left), toBytes(right)...)
	return hasher.hashfunc(append(buffer, 0x01))
}

// creates a new Merkle tree with the provided hasher
func New(hashfunc HashFunc) *Tree {
	hasher := Hasher{hashfunc}
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
