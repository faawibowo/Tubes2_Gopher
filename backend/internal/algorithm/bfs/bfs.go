package bfs

import (
	"sync"
	"time"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Queue"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

type BFSResult struct {
	Tree            *Tree.Tree
	NodeCount       int
	CompletePaths   int
	ExecutionTimeMs int64
}

func BuildRecipeTree(target *Graph.ElementGraph, graphMap map[string]*Graph.ElementGraph, maxCount int) BFSResult {
	start := time.Now()

	root := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: root}

	var nodeCount int = 0
	var pathCount int = 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	bfsMultiThreading(target, root, graphMap, &nodeCount, &pathCount, maxCount, &mu, &wg)
	wg.Wait()
	execTimeMs := time.Since(start).Milliseconds()

	pruneIncompletePaths(root, graphMap)

	return BFSResult{
		Tree:            tree,
		NodeCount:       nodeCount,
		CompletePaths:   pathCount,
		ExecutionTimeMs: execTimeMs,
	}
}

func isEqual(a int, b int) bool {
	return a == b
}

func isTreeLeaf(left *Tree.TreeNodeElement, right *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph) bool {
	return Graph.IsLeaf(graphMap[left.Name], graphMap) && Graph.IsLeaf(graphMap[right.Name], graphMap)
}

func bfsMultiThreading(current *Graph.ElementGraph, node *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph, nodeCountPtr *int, pathCountPtr *int, maxCount int, mu *sync.Mutex, wg *sync.WaitGroup) {

	if isEqual(*pathCountPtr, maxCount) {
		return
	}

	q := Queue.NewQueue()
	q.Enqueue(&Queue.QueueItem{GraphNode: current, TreeNode: node})

	for !q.IsEmpty() {
		queueSize := q.Size()
		for i := 0; i < queueSize; i++ {
			item := q.Dequeue()
			if item == nil {
				continue
			}

			current := item.GraphNode
			node := item.TreeNode

			mu.Lock()
			(*nodeCountPtr)++
			mu.Unlock()

			for _, recipe := range current.Recipes {

				if recipe.FirstElement.Tier > current.Tier || recipe.SecondElement.Tier > current.Tier {
					continue
				}

				wg.Add(1)
				go func(r Graph.Recipe, parent *Tree.TreeNodeElement) {
					defer wg.Done()

					left := &Tree.TreeNodeElement{Name: r.FirstElement.Name}
					right := &Tree.TreeNodeElement{Name: r.SecondElement.Name}
					recipeNode := Tree.TreeNodeRecipe{
						FirstElement:  left,
						SecondElement: right,
						ResultElement: parent,
					}
					left.Parent = &recipeNode
					right.Parent = &recipeNode

					mu.Lock()
					if isTreeLeaf(left, right, graphMap) {
						(*pathCountPtr)++
					}
					if *pathCountPtr > maxCount {
						(*pathCountPtr)--
						mu.Unlock()
						return
					}
					parent.Children = append(parent.Children, recipeNode)

					if isEqual(*pathCountPtr, maxCount) {
						mu.Unlock()
						return
					}
					mu.Unlock()

					q.Enqueue(&Queue.QueueItem{GraphNode: r.FirstElement, TreeNode: left})
					q.Enqueue(&Queue.QueueItem{GraphNode: r.SecondElement, TreeNode: right})
				}(recipe, node)
			}
			wg.Wait()
		}
	}
}

func pruneIncompletePaths(node *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph) {
	if node == nil {
		return
	}

	validRecipes := make([]Tree.TreeNodeRecipe, 0, len(node.Children))

	for _, recipe := range node.Children {
		pruneIncompletePaths(recipe.FirstElement, graphMap)
		pruneIncompletePaths(recipe.SecondElement, graphMap)

		if len(recipe.FirstElement.Children) == 0 && len(recipe.SecondElement.Children) == 0 {
			leaf1, leaf2 := false, false
			if elem, ok := graphMap[recipe.FirstElement.Name]; ok {
				leaf1 = Graph.IsLeaf(elem, graphMap)
			}
			if elem, ok := graphMap[recipe.SecondElement.Name]; ok {
				leaf2 = Graph.IsLeaf(elem, graphMap)
			}

			if leaf1 && leaf2 {
				validRecipes = append(validRecipes, recipe)
			}
		} else {
			validRecipes = append(validRecipes, recipe)
		}
	}
	node.Children = validRecipes
}

//================================ Find First Path ====================================

func FindShortestPath(target *Graph.ElementGraph, graphMap map[string]*Graph.ElementGraph) BFSResult {
	start := time.Now()

	root := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: root}
	nodeCount := 0
	found := false

	bfsSequential(target, root, graphMap, &nodeCount, &found)
	pruneIncompletePaths(root, graphMap)

	execTimeMs := time.Since(start).Milliseconds()
	if !found {
		return BFSResult{Tree: tree, NodeCount: nodeCount, CompletePaths: 0, ExecutionTimeMs: execTimeMs}
	}
	return BFSResult{Tree: tree, NodeCount: nodeCount, CompletePaths: 1, ExecutionTimeMs: execTimeMs}
}

// bfsSequential processes the BFS level‐by‐level without launching goroutines.
// It expands each node’s recipes and terminates only if a complete (solved) path
// is found—that is, when the chain of recipes ends in basic elements.
func bfsSequential(target *Graph.ElementGraph, root *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph, nodeCount *int, found *bool) {
	type bfsItem struct {
		elem *Graph.ElementGraph
		node *Tree.TreeNodeElement
	}

	// Initialize queue with the target element.
	queue := []bfsItem{{target, root}}

	for len(queue) > 0 && !*found {
		var next []bfsItem

		// Process all items in the current level.
		for _, itm := range queue {
			// Increment node counter.
			*nodeCount++

			// If the element has no recipes, nothing to expand.
			if len(itm.elem.Recipes) == 0 {
				continue
			}

			parentTier := itm.elem.Tier

			// Expand each valid recipe.
			for _, recipe := range itm.elem.Recipes {
				// Skip recipe if either ingredient's Tier exceeds parent's Tier.
				if recipe.FirstElement.Tier > parentTier || recipe.SecondElement.Tier > parentTier {
					continue
				}

				// Create new tree nodes for the recipe ingredients.
				left := &Tree.TreeNodeElement{Name: recipe.FirstElement.Name}
				right := &Tree.TreeNodeElement{Name: recipe.SecondElement.Name}
				recipeNode := Tree.TreeNodeRecipe{
					FirstElement:  left,
					SecondElement: right,
					ResultElement: itm.node,
				}
				left.Parent = &recipeNode
				right.Parent = &recipeNode

				// Append this recipe to the parent's children.
				itm.node.Children = append(itm.node.Children, recipeNode)

				// Check if the full path from this recipe upward is solved.
				if checkFullPath(&recipeNode) {
					*found = true
					return
				}

				// Enqueue non-basic elements for next level expansion.
				if !isBasic(recipe.FirstElement) {
					next = append(next, bfsItem{recipe.FirstElement, left})
				}
				if !isBasic(recipe.SecondElement) {
					next = append(next, bfsItem{recipe.SecondElement, right})
				}
			}
		}
		queue = next
	}
}

// checkFullPath traverses upward from a recipe node.
// It returns true only if every recipe in the chain is "solved" (i.e. both children are basic or eventually solved).
func checkFullPath(r *Tree.TreeNodeRecipe) bool {
	curr := r
	for curr != nil {
		if !isRecipeSolved(curr) {
			return false
		}
		// If we've reached the root (no parent), the chain is complete.
		if curr.ResultElement.Parent == nil {
			return true
		}
		curr = curr.ResultElement.Parent
	}
	return false
}

// isRecipeSolved returns true if both children of a recipe are solved.
func isRecipeSolved(r *Tree.TreeNodeRecipe) bool {
	return isSolvedElem(r.FirstElement) && isSolvedElem(r.SecondElement)
}

// A node is solved if it has no children and its name is one of the basic elements,
// or if at least one of its recipe branches is solved.
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

// isBasic returns true if the given graph element is a basic element.
func isBasic(e *Graph.ElementGraph) bool {
	basic := map[string]bool{
		"Air":   true,
		"Water": true,
		"Earth": true,
		"Fire":  true,
	}
	return basic[e.Name]
}

// IsLeafName returns true if the provided name corresponds to a basic element.
func IsLeafName(name string) bool {
	return name == "Air" || name == "Earth" || name == "Water" || name == "Fire"
}
