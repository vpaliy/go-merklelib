package merklelib

type NodePosition int16

const (
	Undefined NodePosition = iota + 1
	Right
	Left
)

// TODO: make it private?
type Positioner interface {
	position() NodePosition
}

// Building block
type MerkleNode struct {
	hashVal             []byte
	left, right, parent *MerkleNode
}

type AuditProofNode struct {
	hashVal []byte
	pos     NodePosition
}

var sentinel = new(MerkleNode)

func newNode(hashVal []byte) *MerkleNode {
	return &MerkleNode{hashVal, nil, nil, nil}
}

func (node *MerkleNode) position() NodePosition {
	switch {
	case node.parent == nil:
		return Undefined
	case node.parent.left == node:
		return Left
	default:
		return Right
	}
}

func (node *AuditProofNode) position() NodePosition {
	return node.pos
}

func getHashVal(node interface{}) []byte {
	switch value := node.(type) {
	case *MerkleNode:
		return value.hashVal
	case *AuditProofNode:
		return value.hashVal
	case AuditProofNode:
		return value.hashVal
	case []byte:
		return value
	default:
		// TODO: add error to the return type?
		// I can't see a scenario when this could happen
		return nil
	}
}

func isSentinel(node interface{}) bool {
	if value, ok := node.(*MerkleNode); ok {
		return value == sentinel
	}
	return false
}

func isRight(node interface{}) bool {
	if value, ok := node.(Positioner); ok {
		return value.position() == Right
	}
	return false
}

func isLeft(node interface{}) bool {
	if value, ok := node.(Positioner); ok {
		return value.position() == Left
	}
	return false
}

func concat(hasher *Hasher, left, right interface{}) []byte {
	if isSentinel(right) {
		return getHashVal(left)
	} else if isSentinel(left) {
		return getHashVal(right)
	}

	if isRight(left) || isLeft(right) {
		return hasher.HashChildren(right, left)
	}

	return hasher.HashChildren(left, right)
}

func mergeNodes(hasher *Hasher, left, right *MerkleNode) *MerkleNode {
	hashVal := concat(hasher, left, right)
	newNode := &MerkleNode{hashVal, left, right, nil}
	// update their parents
	right.parent = newNode
	left.parent = newNode
	return newNode
}

func (node *MerkleNode) sibling() *MerkleNode {
	parent := node.parent
	switch {
	case parent == nil:
		return nil
	case parent.left == node:
		return parent.right
	default:
		return parent.left
	}
}

func (m *MerkleNode) climbTo(level int) *MerkleNode {
	node := m
	for ; level > 0; level-- {
		if node == nil {
			return nil
		}
		node = node.parent
	}
	return node
}

func (node *MerkleNode) createAuditProofNode() *AuditProofNode {
	a := AuditProofNode{node.hashVal, node.position()}
	return &a
}
