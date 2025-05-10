package bidirectional

import (
	"runtime"
	"sync"
	"time"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

var basic = map[string]bool{
	"Air": true, "Water": true, "Earth": true, "Fire": true,
}

func isBasic(e *Graph.ElementGraph) bool { return basic[e.Name] }

func BuildRecipeTreeBFSConcurrent(
	target *Graph.ElementGraph,
	maxPaths int,
	delay time.Duration,
	onStep func(*Tree.TreeNodeRecipe),
	_ map[string]*Graph.ElementGraph, // graphMap kept in signature, but unused
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
		done      bool // ★ new : global stop-flag
	)

	checkFullPath := func(r *Tree.TreeNodeRecipe) bool { // ★ new
		// walk upward until root; every step must be solved
		curr := r
		for curr != nil {
			if !isRecipeSolved(curr) {
				return false
			}
			if curr.ResultElement.Parent == nil {
				return true // reached root; whole path solved
			}
			curr = curr.ResultElement.Parent
		}
		return false
	}

	//------------------------------------------------------------------
	queue := []bfsItem{{target, root}}

	for len(queue) > 0 && !done {
		var next []bfsItem

		for _, itm := range queue {
			parentTier := itm.elem.Tier
			if len(itm.elem.Recipes) == 0 {
				continue
			}

			for _, recipe := range itm.elem.Recipes {
				// Check if done already
				mu.Lock()
				if done {
					mu.Unlock()
					break
				}
				mu.Unlock()

				// ─── Tier gate ──────────────────────────────────────────
				if recipe.FirstElement.Tier > parentTier ||
					recipe.SecondElement.Tier > parentTier {
					continue
				}
				// ────────────────────────────────────────────────────────

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

					// ★ link upward
					left.Parent = &step
					right.Parent = &step

					time.Sleep(delay)

					mu.Lock()
					defer mu.Unlock()

					if done {
						return
					}

					parent.Children = append(parent.Children, step)

					if onStep != nil {
						onStep(&step)
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

	Tree.PruneTreeToBasicPaths(tree.First)
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
