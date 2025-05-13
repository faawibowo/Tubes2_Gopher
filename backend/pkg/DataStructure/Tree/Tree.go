package Tree

import (
	"fmt"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
)

type Tree struct {
	First *TreeNodeElement `json:"first"`
}

type TreeNodeElement struct {
	Name     string           `json:"name"`
	Children []TreeNodeRecipe `json:"children"`
	Parent   *TreeNodeRecipe  `json:"-"` // ← ini JANGAN ikut json biar gak circular
}

type TreeNodeRecipe struct {
	FirstElement  *TreeNodeElement `json:"firstElement"`
	SecondElement *TreeNodeElement `json:"secondElement"`
	ResultElement *TreeNodeElement `json:"-"` // remove from JSON!
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

func PrintTree(node *TreeNodeElement, depth int, printName bool, prefixStack []bool) {
	if node == nil {
		return
	}
	prefix := ""
	for _, active := range prefixStack {
		if active {
			prefix += "│  "
		} else {
			prefix += "   "
		}
	}

	if printName {
		fmt.Println(prefix + node.Name)
	}

	for i, recipe := range node.Children {
		isLast := i == len(node.Children)-1

		recipePrefix := prefix + "├─ Recipe:"
		if isLast {
			recipePrefix = prefix + "└─ Recipe:"
		}
		fmt.Println(recipePrefix)

		childStack := append(prefixStack, !isLast)

		leftPrefix := ""
		for _, active := range childStack {
			if active {
				leftPrefix += "│  "
			} else {
				leftPrefix += "   "
			}
		}
		fmt.Println(leftPrefix + "├─ " + recipe.FirstElement.Name + " (left)")
		PrintTree(recipe.FirstElement, depth+1, false, append(childStack, true))

		rightPrefix := ""
		for _, active := range childStack {
			if active {
				rightPrefix += "│  "
			} else {
				rightPrefix += "   "
			}
		}
		fmt.Println(rightPrefix + "└─ " + recipe.SecondElement.Name + " (right)")
		PrintTree(recipe.SecondElement, depth+1, false, append(childStack, false))
	}
}

func IsBasic(name string) bool {
	return name == "Air" || name == "Water" || name == "Earth" || name == "Fire"
}

func CountPaths(node *TreeNodeElement) int {
	if node == nil {
		return 0
	}
	if len(node.Children) == 0 {
		return 1
	}

	total := 0
	for _, recipe := range node.Children {
		left := CountPaths(recipe.FirstElement)
		right := CountPaths(recipe.SecondElement)
		total += left * right
	}
	return total
}
