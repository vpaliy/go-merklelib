package merklelib

type NodeType int16

const (
  Left  NodeType = 0
  Right NodeType = 1
  Leaf  NodeType = 2
)

// Building block
type MerkleNode struct {
  hashVal []byte
  left, right, parent *MerkleNode
}

var sentinel = new(MerkleNode)

func newNode(hashVal []byte) *MerkleNode {
  return &MerkleNode { hashVal, nil, nil, nil }
}

func (node *MerkleNode) Position() NodeType {
  switch  {
  case node.parent == nil:
    return Leaf
  case node.parent.left == node:
    return Left
  default:
    return Right
  }
}

func concatHashes(hasher Hasher, left, right *MerkleNode) [] byte {
  switch  {
  case left == sentinel:
    return right.hashVal
  case right == sentinel:
    return left.hashVal
  case left.Position() == Right || right.Position() == Left:
    return hasher.HashChildren(right, left)
  default:
    return hasher.HashChildren(left, right)
  }
}

func mergeNodes(hasher Hasher, left, right *MerkleNode) *MerkleNode {
  hashVal := concatHashes(hasher, left, right)
  newNode := &MerkleNode { hashVal, left, right, nil }
  // update their parents
  right.parent = newNode
  left.parent = newNode
  return newNode
}
