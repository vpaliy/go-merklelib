package merklelib

type node struct {
  key string
  next, prev *node
}

type OrderedMap struct {
  head, tail *node
  mapping map[string]*MerkleNode
}

func (dict *OrderedMap) Add(key string, value *MerkleNode) {
  newNode := &node {key, nil, nil }
  if dict.head == nil {
    dict.head = newNode
    dict.tail = newNode
  } else {
    newNode.prev = dict.tail
    dict.tail.next = newNode
    dict.tail = newNode
  }
  dict.mapping[key] = value
}

func (dict *OrderedMap) Remove(key string) {
  // TODO:
}

func (dict *OrderedMap) Get(key string) *MerkleNode {
  return dict.mapping[key]
}

func (dict *OrderedMap) Keys() []string {
  if dict.head == nil {
    return nil
  }
  keys := make([]string, len(dict.mapping))
  root := dict.head
  for i := 0; root != nil; i++ {
    keys[i] = root.key
  }
  return keys
}

func (dict *OrderedMap) Values() []*MerkleNode {
  if dict.head == nil {
    return nil
  }
  root, mapping := dict.head, dict.mapping
  values := make([]*MerkleNode, len(mapping))
  for i := 0; root != nil; i++ {
    values[i] = mapping[root.key]
  }
  return values
}

func (dict *OrderedMap) Last() *MerkleNode {
  if dict.head == nil {
    return nil
  }
  return dict.mapping[dict.tail.key]
}

func (dict *OrderedMap) First() *MerkleNode {
  if dict.head == nil {
    return nil
  }
  return dict.mapping[dict.head.key]
}
