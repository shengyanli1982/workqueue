package main

import (
	"fmt"

	lst "github.com/shengyanli1982/workqueue/v2/internal/container/list"
	memalign "github.com/vearne/mem-align"
)

func main() {
	fmt.Printf("Node struct alignment:\n\n")
	memalign.PrintStructAlignment(lst.Node{})
}
