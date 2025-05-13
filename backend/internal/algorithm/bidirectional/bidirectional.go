package bidirectional

import (
	"runtime"
	"sync"
	"time"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

type TreeUpdate struct {
	Done bool       `json:"done"`
	Tree *Tree.Tree `json:"tree"`
}

var basic = map[string]bool{
	"Air": true, "Water": true, "Earth": true, "Fire": true,
}

func isBasic(e *Graph.ElementGraph) bool { return basic[e.Name] }

func BuildRecipeTreeBFSConcurrent(
	target *Graph.ElementGraph,
	maxPaths int,
	delay time.Duration,
	updates chan<- *TreeUpdate,
	_ map[string]*Graph.ElementGraph,
) *Tree.Tree {

	root := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: root}

	type bfsItem struct {
		elem *Graph.ElementGraph
		node *Tree.TreeNodeElement
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		pathCount int
		done      bool
	)

	checkFullPath := func(r *Tree.TreeNodeRecipe) bool {
		curr := r
		for curr != nil {
			if !isRecipeSolved(curr) {
				return false
			}
			if curr.ResultElement.Parent == nil {
				return true
			}
			curr = curr.ResultElement.Parent
		}
		return false
	}

	queue := []bfsItem{{target, root}}

	for len(queue) > 0 && !done {
		var next []bfsItem

		for _, itm := range queue {
			parentTier := itm.elem.Tier
			if len(itm.elem.Recipes) == 0 {
				continue
			}

			for _, recipe := range itm.elem.Recipes {
				mu.Lock()
				if done {
					mu.Unlock()
					break
				}
				mu.Unlock()

				if recipe.FirstElement.Tier > parentTier ||
					recipe.SecondElement.Tier > parentTier {
					continue
				}

				wg.Add(1)
				go func(rc Graph.Recipe, parent *Tree.TreeNodeElement) {
					defer wg.Done()

					left := &Tree.TreeNodeElement{Name: rc.FirstElement.Name}
					right := &Tree.TreeNodeElement{Name: rc.SecondElement.Name}
					step := Tree.TreeNodeRecipe{
						FirstElement:  left,
						SecondElement: right,
						ResultElement: parent,
					}

					left.Parent = &step
					right.Parent = &step

					time.Sleep(delay)

					mu.Lock()
					defer mu.Unlock()

					if done {
						return
					}

					parent.Children = append(parent.Children, step)

					if updates != nil {
						cloned := Tree.CopyTree(tree)
						updates <- &TreeUpdate{
							Done: false,
							Tree: cloned,
						}
					}

					if !isBasic(rc.FirstElement) {
						next = append(next, bfsItem{rc.FirstElement, left})
					}
					if !isBasic(rc.SecondElement) {
						next = append(next, bfsItem{rc.SecondElement, right})
					}

					if checkFullPath(&step) {
						pathCount++
						if maxPaths > 0 && pathCount >= maxPaths {
							done = true
						}
					}
				}(recipe, itm.node)
			}
		}

		wg.Wait()
		queue = next
	}

	PruneTreeToBasicPaths(tree.First)
	return tree
}

func isRecipeSolved(r *Tree.TreeNodeRecipe) bool {
	return isSolvedElem(r.FirstElement) && isSolvedElem(r.SecondElement)
}

func isSolvedElem(n *Tree.TreeNodeElement) bool {
	if len(n.Children) == 0 {
		return IsLeafName(n.Name)
	}
	for _, c := range n.Children {
		if isRecipeSolved(&c) {
			return true
		}
	}
	return false
}

func IsLeafName(name string) bool {
	return name == "Air" || name == "Earth" || name == "Water" || name == "Fire"
}

func PruneTreeToBasicPaths(n *Tree.TreeNodeElement) bool {
	if n == nil {
		return false
	}

	if Tree.IsBasic(n.Name) {
		return true
	}

	valid := make([]Tree.TreeNodeRecipe, 0, len(n.Children))

	for _, rc := range n.Children {

		leftValid := PruneTreeToBasicPaths(rc.FirstElement)
		rightValid := PruneTreeToBasicPaths(rc.SecondElement)

		if leftValid && rightValid {

			valid = append(valid, rc)
		}
	}

	n.Children = valid

	return len(valid) > 0
}
