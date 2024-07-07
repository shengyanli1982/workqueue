package set

// setValue 是一个空结构体，用于 Set 中 map 的值。
// setValue is an empty struct, used as the value in the map of Set.
var setValue = struct{}{}

// Set 是一个结构体，包含一个 map，用于存储集合的元素。
// Set is a struct, containing a map, used to store the elements of the set.
type Set struct {
	m map[interface{}]struct{}
}

// New 函数创建并返回一个新的 Set。
// The New function creates and returns a new Set.
func New() *Set {
	return &Set{
		// 初始化 map。
		// Initialize the map.
		m: make(map[interface{}]struct{}),
	}
}

// Add 方法将一个元素添加到 Set 中。
// The Add method adds an element to the Set.
func (s *Set) Add(item interface{}) {
	// 在 map 中添加一个键值对，键是 item，值是 setValue。
	// Add a key-value pair to the map, the key is item, the value is setValue.
	s.m[item] = setValue
}

// Remove 方法从 Set 中移除一个元素。
// The Remove method removes an element from the Set.
func (s *Set) Remove(item interface{}) {
	// 在 map 中删除键为 item 的键值对。
	// Delete the key-value pair with key item in the map.
	delete(s.m, item)
}

// Contains 方法检查一个元素是否在 Set 中。
// The Contains method checks whether an element is in the Set.
func (s *Set) Contains(item interface{}) bool {
	// 在 map 中查找键为 item 的键值对，如果找到，返回 true，否则返回 false。
	// Find the key-value pair with key item in the map, return true if found, otherwise return false.
	_, c := s.m[item]
	return c
}

// Len 方法返回 Set 的元素个数。
// The Len method returns the number of elements in the Set.
func (s *Set) Len() int {
	// 返回 map 的键值对个数，即 Set 的元素个数。
	// Return the number of key-value pairs in the map, i.e., the number of elements in the Set.
	return len(s.m)
}

// List 方法返回 Set 的所有元素的列表。
// The List method returns a list of all elements in the Set.
func (s *Set) List() []interface{} {
	// 创建一个空的列表，长度为 map 的键值对个数。
	// Create an empty list, the length is the number of key-value pairs in the map.
	list := make([]interface{}, 0, len(s.m))

	// 遍历 map 的所有键，将键添加到列表中。
	// Iterate over all keys in the map, add the keys to the list.
	for item := range s.m {
		list = append(list, item)
	}

	// 返回列表。
	// Return the list.
	return list
}

// Cleanup 方法清空 Set。
// The Cleanup method clears the Set.
func (s *Set) Cleanup() {
	// 创建一个新的空 map，赋值给 Set 的 map，原来的 map 将被垃圾回收。
	// Create a new empty map, assign it to the map of Set, the original map will be garbage collected.
	s.m = make(map[interface{}]struct{})
}
