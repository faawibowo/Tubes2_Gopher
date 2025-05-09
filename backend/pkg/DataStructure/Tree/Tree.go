package Tree

import "github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"

type Tree struct {
	First *TreeNodeElement
}

type TreeNodeElement struct {
	Name     string
	Parent   *TreeNodeRecipe
	Children []TreeNodeRecipe
}

type TreeNodeRecipe struct {
	FirstElement  *TreeNodeElement
	SecondElement *TreeNodeElement
	ResultElement *TreeNodeElement
}

// NewTree : buat pohon dengan satu root
func NewTree(node *Graph.ElementGraph) *Tree {

	root := &TreeNodeElement{Name: node.Name}
	return &Tree{First: root}

}

// MakeTree : tambahkan satu kombinasi (recipe) ke node induk
func (t *Tree) MakeTree(
	parent *TreeNodeElement,
	firstChild *Graph.ElementGraph,
	secondChild *Graph.ElementGraph,
) {
	if parent == nil || firstChild == nil || secondChild == nil {
		return
	}

	left := &TreeNodeElement{Name: firstChild.Name}
	right := &TreeNodeElement{Name: secondChild.Name}

	recipe := TreeNodeRecipe{
		FirstElement:  left,
		SecondElement: right,
		ResultElement: parent,
	}

	left.Parent, right.Parent = &recipe, &recipe
	parent.Children = append(parent.Children, recipe)
}

// ----------  utilitas  ----------

// CountNodes : hitung total node elemen dalam tree
func CountNodes(t *Tree) int {
	if t == nil || t.First == nil {
		return 0
	}
	return countElem(t.First)
}

func countElem(n *TreeNodeElement) int {
	if n == nil {
		return 0
	}
	total := 1 // node ini
	for _, childR := range n.Children {
		total += countElem(childR.FirstElement)
		total += countElem(childR.SecondElement)
	}
	return total
}

func CopyTree(src *Tree) *Tree {
	if src == nil || src.First == nil {
		return nil
	}
	return &Tree{First: copyElem(src.First, nil)}
}

func copyElem(orig *TreeNodeElement, parentRecipe *TreeNodeRecipe) *TreeNodeElement {
	if orig == nil {
		return nil
	}
	newNode := &TreeNodeElement{
		Name:   orig.Name,
		Parent: parentRecipe, // bisa nil untuk root
	}
	for _, rc := range orig.Children {
		// buat salinan recipe & anak2
		left := copyElem(rc.FirstElement, nil)
		right := copyElem(rc.SecondElement, nil)
		newRc := TreeNodeRecipe{
			FirstElement:  left,
			SecondElement: right,
			ResultElement: newNode,
		}
		left.Parent, right.Parent = &newRc, &newRc
		newNode.Children = append(newNode.Children, newRc)
	}
	return newNode
}
