package lrucache

// List ...
type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

// ListItem ...
type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	size int
	head *ListItem
	tail *ListItem
}

// NewList ..
func NewList() List {
	return new(list)
}

func (ll *list) Len() int {
	return ll.size
}

func (ll *list) Back() *ListItem {
	return ll.tail
}

func (ll *list) Front() *ListItem {
	return ll.head
}

func (ll *list) PushBack(v interface{}) *ListItem {
	if v == nil {
		return ll.tail
	}
	ll.tail = &ListItem{v, nil, ll.tail}
	if ll.tail.Prev != nil {
		ll.tail.Prev.Next = ll.tail
	}

	if ll.head == nil {
		ll.head = ll.tail
	}
	ll.size++

	return ll.tail
}

func (ll *list) PushFront(v interface{}) *ListItem {
	if v == nil {
		return ll.head
	}
	ll.head = &ListItem{v, ll.head, nil}
	if ll.head.Next != nil {
		ll.head.Next.Prev = ll.head
	}
	if ll.tail == nil {
		ll.tail = ll.head
	}
	ll.size++

	return ll.head
}

func (ll *list) Remove(i *ListItem) {
	if i == nil {
		return
	}
	ll.size--
	defer ll.clearItem(i)
	if i.Next == nil && i.Prev == nil {
		return
	}
	if i.Next == nil {
		ll.tail = ll.tail.Prev
		ll.tail.Next = nil
		return
	}
	if i.Prev == nil {
		ll.head = ll.head.Next
		ll.head.Prev = nil
		return
	}
	i.Next.Prev, i.Prev.Next = i.Prev, i.Next
}

func (ll *list) MoveToFront(i *ListItem) {
	if i == nil {
		return
	}
	if i.Next == nil && i.Prev == nil {
		return
	}
	if i.Prev == nil {
		return
	}
	if i.Next == nil {
		ll.tail = ll.tail.Prev
		ll.tail.Next = nil
		ll.setFront(i)
		return
	}
	i.Next.Prev, i.Prev.Next = i.Prev, i.Next
	ll.setFront(i)
}

func (ll *list) clearItem(i *ListItem) {
	i.Prev = nil
	i.Next = nil
	i.Value = nil
}

func (ll *list) setFront(v *ListItem) {
	v.Next = ll.head
	v.Prev = nil
	ll.head.Prev = v
	ll.head = v
}
