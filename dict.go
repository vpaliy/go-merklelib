package merklelib

type mapnode struct {
	key        string
	value      *MerkleNode
	next, prev *mapnode
}

type OrderedMap struct {
	head, tail *mapnode
	mapping    map[string]*mapnode
}

func (dict *OrderedMap) Add(key string, value *MerkleNode) {
	newNode := &mapnode{key, value, nil, nil}
	if dict.head == nil {
		dict.head = newNode
		dict.tail = newNode
	} else {
		newNode.prev = dict.tail
		dict.tail.next = newNode
		dict.tail = newNode
	}
	dict.mapping[key] = newNode
}

func (dict *OrderedMap) Remove(key string) {
	node := dict.mapping[key]

	if dict.head == node {
		dict.head = dict.head.next
		if dict.head != nil {
			dict.head.prev = nil
		}
	}

	if dict.tail == node {
		dict.tail = dict.tail.prev
		if dict.tail != nil {
			dict.tail.next = nil
		}
	}

	delete(dict.mapping, key)
}

func (dict *OrderedMap) Get(key string) *MerkleNode {
	node := dict.mapping[key]
	return node.value
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
		values[i] = dict.Get(root.key)
	}
	return values
}

func (dict *OrderedMap) Last() *MerkleNode {
	if dict.head == nil {
		return nil
	}
	return dict.Get(dict.tail.key)
}

func (dict *OrderedMap) First() *MerkleNode {
	if dict.head == nil {
		return nil
	}
	return dict.Get(dict.head.key)
}

func (dict *OrderedMap) Clear() {
	for key := range dict.mapping {
		delete(dict.mapping, key)
	}

	node := dict.head
	for node != nil {
		node.value = nil
		if node.prev != nil {
			node.prev.next = nil
			node.prev = nil
		}
		node = node.next
	}
}
