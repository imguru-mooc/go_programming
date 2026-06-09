package main

import "fmt"

func main() {
	var list []int
	list = append(list, 1)
	list = append(list, 2)
	list = append(list, 3)
	// 또는: list := []int{1, 2, 3}

	// 출력
	for i, v := range list {
		if i > 0 {
			fmt.Print(" -> ")
		}
		fmt.Print(v)
	}
	fmt.Println()

	// free 같은 거 안 함 — GC가 알아서
}

/*
package main

import "fmt"

type Node struct {
	value int
	next  *Node
}

func (head *Node) push(v int) *Node {
	n := new(Node)
	n.value = v
	n.next = head
	return n
}

func (head *Node) print_list() {
	for head != nil {
		fmt.Printf("%d -> ", head.value)
		head = head.next
	}
	fmt.Println("NULL")
}

func (head *Node) free_list() {
	for head != nil {
		next := head.next
		head = next
	}
}

func main() {
	var head *Node = nil
	head = head.push(1)
	head = head.push(2)
	head = head.push(3)
	head.print_list()
	head.free_list()
	head = nil
}
*/

/*
package main

import "fmt"

type Node struct {
	value int
	next  *Node
}

func push(head *Node, v int) *Node {
	n := new(Node)
	n.value = v
	n.next = head
	return n
}

func print_list(head *Node) {
	for head != nil {
		fmt.Printf("%d -> ", head.value)
		head = head.next
	}
	fmt.Println("NULL")
}

func free_list(head *Node) {
	for head != nil {
		next := head.next
		head = next
	}
}

func main() {
	var head *Node = nil
	head = push(head, 1)
	head = push(head, 2)
	head = push(head, 3)
	print_list(head)
	free_list(head)
	head = nil
}
*/
