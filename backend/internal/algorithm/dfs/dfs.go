package dfs

import (
	"sync"
	"time"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

type DFSResult struct {
	Tree            *Tree.Tree `json:"tree"`
	NodeCount       int        `json:"nodeCount"`
	CompletePaths   int        `json:"completePaths"`
	ExecutionTimeMs int64      `json:"executionTimeMs"`
	Done            bool       `json:"done"`
}

func BuildRecipeTree(
	target *Graph.ElementGraph,
	graphMap map[string]*Graph.ElementGraph,
	maxCount int,
	delay time.Duration,
	updates chan<- DFSResult,
) DFSResult {
	start := time.Now()

	root := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: root}

	var nodeCount int = 0
	var pathCount int = 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	dfsMultiThreading(target, root, graphMap, &nodeCount, &pathCount, maxCount, delay, updates, &mu, &wg, tree, start)
	wg.Wait()

	pruneIncompletePaths(root, graphMap)
	wg.Wait()

	return DFSResult{
		Tree:            tree,
		NodeCount:       nodeCount,
		CompletePaths:   pathCount,
		ExecutionTimeMs: time.Since(start).Milliseconds(),
		Done:            true,
	}
}

func isEqual(a int, b int) bool {
	return a == b
}

func isTreeLeaf(left *Tree.TreeNodeElement, right *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph) bool {
	return Graph.IsLeaf(graphMap[left.Name], graphMap) && Graph.IsLeaf(graphMap[right.Name], graphMap)
}

func dfsMultiThreading(
	current *Graph.ElementGraph,
	node *Tree.TreeNodeElement,
	graphMap map[string]*Graph.ElementGraph,
	nodeCountPtr *int,
	pathCountPtr *int,
	maxCount int,
	delay time.Duration,
	updates chan<- DFSResult,
	mu *sync.Mutex,
	wg *sync.WaitGroup,
	tree *Tree.Tree,
	start time.Time,
) {
	if isEqual(*pathCountPtr, maxCount) {
		return
	}

	mu.Lock()
	(*nodeCountPtr)++
	mu.Unlock()

	for _, recipe := range current.Recipes {
		if isEqual(*pathCountPtr, maxCount) {
			return
		}

		wg.Add(1)
		time.Sleep(delay)
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
			if isEqual(*pathCountPtr, maxCount) {
				mu.Unlock()
				return
			}
			node.Children = append(node.Children, recipeNode)
			if isTreeLeaf(left, right, graphMap) {
				(*pathCountPtr)++
			}
			mu.Unlock()

			left.Parent = &recipeNode
			right.Parent = &recipeNode

			// ✨ Send update
			if updates != nil {
				cloned := Tree.CopyTree(tree)
				updates <- DFSResult{
					Tree:            cloned,
					NodeCount:       *nodeCountPtr,
					CompletePaths:   *pathCountPtr,
					ExecutionTimeMs: time.Since(start).Milliseconds(),
					Done:            false,
				}
			}

			if r.FirstElement.Tier >= current.Tier || r.SecondElement.Tier >= current.Tier {
				return
			}

			if isEqual(*pathCountPtr, maxCount) {
				return
			}
			dfsMultiThreading(r.FirstElement, left, graphMap, nodeCountPtr, pathCountPtr, maxCount, delay, updates, mu, wg, tree, start)
			dfsMultiThreading(r.SecondElement, right, graphMap, nodeCountPtr, pathCountPtr, maxCount, delay, updates, mu, wg, tree, start)
		}(recipe)
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

func FindShortestPath(
	target *Graph.ElementGraph,
	graphMap map[string]*Graph.ElementGraph,
	delay time.Duration,
	updates chan<- DFSResult,
) DFSResult {
	start := time.Now()

	root := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: root}

	var nodeCount int
	var found bool

	dfsShortest(target, root, graphMap, &nodeCount, delay, updates, tree, start) //  &mu

	return DFSResult{
		Tree:            tree,
		NodeCount:       nodeCount,
		CompletePaths:   boolToInt(found),
		ExecutionTimeMs: time.Since(start).Milliseconds(),
		Done:            true,
	}
}

func dfsShortest(
	current *Graph.ElementGraph,
	node *Tree.TreeNodeElement,
	graphMap map[string]*Graph.ElementGraph,
	nodeCount *int,
	delay time.Duration,
	updates chan<- DFSResult,
	tree *Tree.Tree,
	start time.Time,
) {
	(*nodeCount)++

	for _, recipe := range current.Recipes {
		if recipe.FirstElement.Tier >= current.Tier || recipe.SecondElement.Tier >= current.Tier {
			continue
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

		time.Sleep(delay)

		if updates != nil {
			cloned := Tree.CopyTree(tree)
			updates <- DFSResult{
				Tree:            cloned,
				NodeCount:       *nodeCount,
				CompletePaths:   0,
				ExecutionTimeMs: time.Since(start).Milliseconds(),
				Done:            false,
			}
		}

		dfsShortest(recipe.FirstElement, left, graphMap, nodeCount, delay, updates, tree, start)
		dfsShortest(recipe.SecondElement, right, graphMap, nodeCount, delay, updates, tree, start)

		break
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// usable -> BuildRecipeTree(), FindShortestPath()
