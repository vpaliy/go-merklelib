package merklelib

// an alias for a hash function
type HashFunc func([]byte) []byte

type Hasher struct {
	hashfunc HashFunc
}

type Tree struct {
	hasher Hasher
	root   *MerkleNode
	// TODO: replace this thing
	mapping map[string]*MerkleNode
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
		for index, item := range data {
			hashVal := tree.hasher.HashLeaf(item)
			node := newNode(hashVal)
			tree.mapping[string(hashVal)] = node
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

// append an additional node
func (tree *Tree) Append(item interface{}) {
	if tree.root == nil {
		tree.root = newNode(tree.hasher.HashLeaf(item))
		return
	}

	// TODO: use ordered map here
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
