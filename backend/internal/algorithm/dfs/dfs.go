package dfs

import (
	"sync"
	"time"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

type DFSResult struct {
	Tree            *Tree.Tree
	NodeCount       int
	CompletePaths   int
	ExecutionTimeMs int64
}

func BuildRecipeTree(target *Graph.ElementGraph, graphMap map[string]*Graph.ElementGraph, maxCount int) DFSResult {
	start := time.Now()

	root := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: root}

	var nodeCount, pathCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	dfsMultiThreading(target, root, graphMap, &nodeCount, &pathCount, maxCount, &mu, &wg, false)
	wg.Wait()
	execTimeMs := time.Since(start).Milliseconds()

	removed := pruneIncompletePaths(root, graphMap)

	return DFSResult{
		Tree:            tree,
		NodeCount:       nodeCount - removed,
		CompletePaths:   pathCount,
		ExecutionTimeMs: execTimeMs,
	}
}

func dfsMultiThreading(current *Graph.ElementGraph, node *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph, nodeCountPtr *int, pathCountPtr *int, maxCount int, mu *sync.Mutex, wg *sync.WaitGroup, done bool) {

	mu.Lock()
	(*nodeCountPtr)++
	mu.Unlock()

	if Graph.IsLeaf(current, graphMap) {
		mu.Lock()
		if *pathCountPtr < maxCount {
			*pathCountPtr++
		}
		done = *pathCountPtr == maxCount
		mu.Unlock()
		return
	}

	if done {
		return
	}

	mu.Lock()
	if *pathCountPtr >= maxCount {
		mu.Unlock()
		return
	}
	mu.Unlock()

	for _, recipe := range current.Recipes {
		mu.Lock()
		if *pathCountPtr >= maxCount {
			mu.Unlock()
			break
		}
		mu.Unlock()

		wg.Add(1)
		go func(r Graph.Recipe) {
			defer wg.Done()

			left := &Tree.TreeNodeElement{Name: r.FirstElement.Name}
			right := &Tree.TreeNodeElement{Name: r.SecondElement.Name}

			recipeNode := Tree.TreeNodeRecipe{
				FirstElement:  left,
				SecondElement: right,
				ResultElement: node,
			}

			mu.Lock()
			node.Children = append(node.Children, recipeNode)
			mu.Unlock()

			left.Parent = &recipeNode
			right.Parent = &recipeNode

			dfsMultiThreading(r.FirstElement, left, graphMap, nodeCountPtr, pathCountPtr, maxCount, mu, wg, done)
			dfsMultiThreading(r.SecondElement, right, graphMap, nodeCountPtr, pathCountPtr, maxCount, mu, wg, done)
		}(recipe)
	}
}

func countSubtree(node *Tree.TreeNodeElement) int {
	if node == nil {
		return 0
	}
	count := 1
	for _, recipe := range node.Children {
		count += countSubtree(recipe.FirstElement)
		count += countSubtree(recipe.SecondElement)
	}
	return count
}

func pruneIncompletePaths(node *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph) int {
	if node == nil {
		return 0
	}

	removed := 0
	validRecipes := make([]Tree.TreeNodeRecipe, 0, len(node.Children))

	for _, recipe := range node.Children {
		removed += pruneIncompletePaths(recipe.FirstElement, graphMap)
		removed += pruneIncompletePaths(recipe.SecondElement, graphMap)

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
			} else {
				removed += countSubtree(recipe.FirstElement)
				removed += countSubtree(recipe.SecondElement)
			}
		} else {
			validRecipes = append(validRecipes, recipe)
		}
	}
	node.Children = validRecipes
	return removed
}

func FindShortestPath(target *Graph.ElementGraph, graphMap map[string]*Graph.ElementGraph) DFSResult {
	start := time.Now()

	root := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: root}

	var nodeCount int
	var found bool

	dfsShortest(target, root, graphMap, &nodeCount, &found)

	execTimeMs := time.Since(start).Milliseconds()

	if !found {
		return DFSResult{Tree: tree, NodeCount: nodeCount, CompletePaths: 0, ExecutionTimeMs: execTimeMs}
	}
	return DFSResult{Tree: tree, NodeCount: nodeCount, CompletePaths: 1, ExecutionTimeMs: execTimeMs}
}

func dfsShortest(current *Graph.ElementGraph, node *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph, nodeCount *int, found *bool) {
	if *found {
		return
	}

	*nodeCount++

	if Graph.IsLeaf(current, graphMap) {
		*found = true
		return
	}

	for _, recipe := range current.Recipes {
		if *found {
			break
		}

		left := &Tree.TreeNodeElement{Name: recipe.FirstElement.Name}
		right := &Tree.TreeNodeElement{Name: recipe.SecondElement.Name}

		recipeNode := Tree.TreeNodeRecipe{
			FirstElement:  left,
			SecondElement: right,
			ResultElement: node,
		}

		node.Children = append(node.Children, recipeNode)
		left.Parent = &recipeNode
		right.Parent = &recipeNode

		dfsShortest(recipe.FirstElement, left, graphMap, nodeCount, found)
		dfsShortest(recipe.SecondElement, right, graphMap, nodeCount, found)

		if !*found {
			node.Children = node.Children[:len(node.Children)-1]
		}
	}
}

// usable -> BuildRecipeTree(), FindShortestPath()
