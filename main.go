package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

type node struct {
	val   int32
	left  *node
	right *node
}

type tree struct {
	root  *node
	count int32
}

func (t *tree) add(val int32) {
	atomic.AddInt32(&t.count, 1)
	for {
		root := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&t.root)))
		if root == nil {
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&t.root)), unsafe.Pointer(nil), unsafe.Pointer(&node{val: val, left: nil, right: nil})) {
				return
			}
			continue
		}

		rootBase := (*node)(root)
		rootCopy := node{rootBase.val, rootBase.left, rootBase.right}
		rootCopy.add(val)
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&t.root)), root, unsafe.Pointer(&rootCopy)) {
			return
		}
	}
}

func (n *node) add(val int32) {
	for {
		nodeVal := atomic.LoadInt32(&n.val)
		if nodeVal < val {
			rightPointer := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&n.right)))
			if rightPointer == nil {
				if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&n.right)), unsafe.Pointer(nil), unsafe.Pointer(&node{val, nil, nil})) {
					return
				}
				continue
			}
			rightBase := (*node)(rightPointer)
			rightCopy := node{rightBase.val, rightBase.left, rightBase.right}
			rightCopy.add(val)
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&n.right)), rightPointer, unsafe.Pointer(&rightCopy)) {
				return
			}
		}

		leftPointer := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&n.left)))
		if leftPointer == nil {
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&n.left)), unsafe.Pointer(nil), unsafe.Pointer(&node{val, nil, nil})) {
				return
			}
			continue
		}
		leftBase := (*node)(leftPointer)
		leftCopy := node{leftBase.val, leftBase.left, leftBase.right}
		leftCopy.add(val)
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&n.left)), leftPointer, unsafe.Pointer(&leftCopy)) {
			return
		}
	}
}

func (t *tree) ToSlice() []int32 {
	if atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&t.root))) == nil {
		return []int32{}
	}

	slice := make([]int32, 0, t.count)
	_ = t.root.toSlice(slice, 0)
	return slice[:t.count]
}

func (n *node) toSlice(slice []int32, prevAdded int) int {
	added := 0
	if n.left != nil {
		added += n.left.toSlice(slice, prevAdded)
	}

	slice = slice[:prevAdded+added]
	slice = append(slice, n.val)
	added++

	if n.right != nil {
		added += n.right.toSlice(slice, prevAdded+added)
	}
	slice = slice[:prevAdded+added]
	return added
}

func newTree() *tree {
	return &tree{}
}

func main() {
	t := newTree()
	wg := sync.WaitGroup{}
	count := int32(0)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			wg.Add(1)
			go func(count int32) {
				defer wg.Done()
				t.add(count)
				fmt.Println("adding: ", count)
			}(count)
			count++
		}
	}
	wg.Wait()
	fmt.Println(t.ToSlice())
}
