package bfs

import (
	"sync"
	"time"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Queue"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

type BFSResult struct {
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
	updates chan<- BFSResult,
) BFSResult {
	start := time.Now()

	root := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: root}

	var nodeCount int = 0
	var pathCount int = 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	bfsMultiThreading(target, root, graphMap, &nodeCount, &pathCount, maxCount, delay, updates, &mu, &wg, tree, start)
	wg.Wait()

	pruneIncompletePaths(root, graphMap)

	return BFSResult{
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

func bfsMultiThreading(
	current *Graph.ElementGraph,
	node *Tree.TreeNodeElement,
	graphMap map[string]*Graph.ElementGraph,
	nodeCountPtr *int,
	pathCountPtr *int,
	maxCount int,
	delay time.Duration,
	updates chan<- BFSResult,
	mu *sync.Mutex,
	wg *sync.WaitGroup,
	tree *Tree.Tree,
	start time.Time,
) {
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
				time.Sleep(delay)
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
					defer mu.Unlock()

					if *pathCountPtr >= maxCount {
						return
					}

					if isTreeLeaf(left, right, graphMap) {
						(*pathCountPtr)++
					}

					parent.Children = append(parent.Children, recipeNode)

					if updates != nil {
						cloned := Tree.CopyTree(tree)
						updates <- BFSResult{
							Tree:            cloned,
							NodeCount:       *nodeCountPtr,
							CompletePaths:   *pathCountPtr,
							ExecutionTimeMs: time.Since(start).Milliseconds(),
							Done:            false,
						}
					}

					if *pathCountPtr >= maxCount {
						return
					}

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

func FindShortestPath(
	target *Graph.ElementGraph,
	graphMap map[string]*Graph.ElementGraph,
	delay time.Duration,
	updates chan<- BFSResult,
) BFSResult {
	start := time.Now()

	root := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: root}

	var nodeCount int
	found := false

	bfsShortest(target, root, graphMap, &nodeCount, &found, delay, updates, tree)
	pruneIncompletePaths(root, graphMap)
	result := BFSResult{
		Tree:            tree,
		NodeCount:       nodeCount,
		CompletePaths:   1,
		ExecutionTimeMs: time.Since(start).Milliseconds(),
		Done:            true,
	}
	if !found {
		result.CompletePaths = 0
	}
	return result
}

func bfsShortest(
	target *Graph.ElementGraph,
	root *Tree.TreeNodeElement,
	graphMap map[string]*Graph.ElementGraph,
	nodeCount *int,
	found *bool,
	delay time.Duration,
	updates chan<- BFSResult,
	tree *Tree.Tree,
) {
	q := Queue.NewQueue()
	q.Enqueue(&Queue.QueueItem{GraphNode: target, TreeNode: root})

	start := time.Now()

	for !q.IsEmpty() && !*found {
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
			step := Tree.TreeNodeRecipe{
				FirstElement:  left,
				SecondElement: right,
				ResultElement: node,
			}

			left.Parent = &step
			right.Parent = &step
			node.Children = append(node.Children, step)

			time.Sleep(delay)

			// ✨ Stream result snapshot
			if updates != nil {
				cloned := Tree.CopyTree(tree)
				updates <- BFSResult{
					Tree:            cloned,
					NodeCount:       *nodeCount,
					CompletePaths:   0,
					ExecutionTimeMs: time.Since(start).Milliseconds(),
					Done:            false,
				}
			}

			if Graph.IsLeaf(graphMap[left.Name], graphMap) && Graph.IsLeaf(graphMap[right.Name], graphMap) {
				*found = true
				return
			}

			q.Enqueue(&Queue.QueueItem{GraphNode: recipe.FirstElement, TreeNode: left})
			q.Enqueue(&Queue.QueueItem{GraphNode: recipe.SecondElement, TreeNode: right})
		}
	}
}
