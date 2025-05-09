package Tree

import (
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
)

type Tree struct {
	Name        string
	FirstChild  *Tree
	SecondChild *Tree
}

func NewTree(node *Graph.Graph) *Tree {
	return &Tree{
		Name: node.Element,
	}
}

func (t *Tree) MakeTree(node *Graph.Graph, firstChild *Graph.Graph, secondChild *Graph.Graph) {
	t.Name = node.Element
	if firstChild != nil {
		t.FirstChild = NewTree(firstChild)
	}
	if secondChild != nil {
		t.SecondChild = NewTree(secondChild)
	}
}

func CountNodes(t *Tree) int {
	if t == nil {
		return 0
	}
	return 1 + CountNodes(t.FirstChild) + CountNodes(t.SecondChild)
}

// Deep copy
func CopyTree(t *Tree) *Tree {
	if t == nil {
		return nil
	}
	newTree := &Tree{
		Name: t.Name,
	}
	newTree.FirstChild = CopyTree(t.FirstChild)
	newTree.SecondChild = CopyTree(t.SecondChild)
	return newTree
}
