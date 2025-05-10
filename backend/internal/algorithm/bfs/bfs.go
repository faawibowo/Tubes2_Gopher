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

func FindShortestPath(target *Graph.ElementGraph, graphMap map[string]*Graph.ElementGraph) BFSResult {
	start := time.Now()

	root := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: root}

	var nodeCount int
	found := false

	bfsShortest(target, root, graphMap, &nodeCount, &found)
	pruneIncompletePaths(root, graphMap)

	execTimeMs := time.Since(start).Milliseconds()

	if !found {
		return BFSResult{Tree: tree, NodeCount: nodeCount, CompletePaths: 0, ExecutionTimeMs: execTimeMs}
	}
	return BFSResult{Tree: tree, NodeCount: nodeCount, CompletePaths: 1, ExecutionTimeMs: execTimeMs}
}

func bfsShortest(target *Graph.ElementGraph, root *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph, nodeCount *int, found *bool) {
	q := Queue.NewQueue()
	q.Enqueue(&Queue.QueueItem{GraphNode: target, TreeNode: root})

	for !q.IsEmpty() {
		item := q.Dequeue()
		if item == nil {
			continue
		}
		current := item.GraphNode
		node := item.TreeNode

		(*nodeCount)++

		for _, recipe := range current.Recipes {
			if recipe.FirstElement.Tier > target.Tier || recipe.SecondElement.Tier > target.Tier {
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

			if isTreeLeaf(left, right, graphMap) {
				*found = true
				return
			}

			q.Enqueue(&Queue.QueueItem{GraphNode: recipe.FirstElement, TreeNode: left})
			q.Enqueue(&Queue.QueueItem{GraphNode: recipe.SecondElement, TreeNode: right})
		}
	}
}
