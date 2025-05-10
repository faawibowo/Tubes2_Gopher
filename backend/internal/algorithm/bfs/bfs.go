package bfs

import (
	"sync"
	"time"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

type BFSResult struct {
	Tree            *Tree.Tree
	NodeCount       int
	CompletePaths   int
	ExecutionTimeMs int64
}

// Queue structure for BFS
type Queue struct {
	items []*QueueItem
	mu    sync.Mutex
}

type QueueItem struct {
	GraphNode *Graph.ElementGraph
	TreeNode  *Tree.TreeNodeElement
}

func NewQueue() *Queue {
	return &Queue{
		items: make([]*QueueItem, 0),
	}
}

func (q *Queue) Enqueue(item *QueueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return nil
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items) == 0
}

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
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

func bfsMultiThreading(target *Graph.ElementGraph, root *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph, nodeCountPtr *int, pathCountPtr *int, maxCount int, mu *sync.Mutex, wg *sync.WaitGroup) {
	if isEqual(*pathCountPtr, maxCount) {
		return
	}
	queue := NewQueue()
	queue.Enqueue(&QueueItem{GraphNode: target, TreeNode: root})

	for !queue.IsEmpty() {
		mu.Lock()
		if isEqual(*pathCountPtr, maxCount) {
			mu.Unlock()
			return
		}
		mu.Unlock()

		currentLevel := queue.Size()
		for i := 0; i < currentLevel; i++ {
			item := queue.Dequeue()
			if item == nil {
				continue
			}

			current := item.GraphNode
			node := item.TreeNode

			mu.Lock()
			(*nodeCountPtr)++
			mu.Unlock()

			if Graph.IsLeaf(current, graphMap) {
				mu.Lock()
				if *pathCountPtr < maxCount {
					*pathCountPtr++
				}
				mu.Unlock()
				continue
			}

			mu.Lock()
			if *pathCountPtr >= maxCount {
				mu.Unlock()
				continue
			}
			mu.Unlock()

			for _, recipe := range current.Recipes {
				if isEqual(*pathCountPtr, maxCount) {
					return
				}

				if recipe.FirstElement.Tier > target.Tier || recipe.SecondElement.Tier > target.Tier {
					continue
				}

				mu.Lock()
				if *nodeCountPtr >= maxCount {
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

					if isEqual(*pathCountPtr, maxCount) {
						return
					}
					queue.Enqueue(&QueueItem{GraphNode: r.FirstElement, TreeNode: left})
					queue.Enqueue(&QueueItem{GraphNode: r.SecondElement, TreeNode: right})
				}(recipe)
			}
			wg.Wait()
		}
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

	execTimeMs := time.Since(start).Milliseconds()

	if !found {
		return BFSResult{Tree: tree, NodeCount: nodeCount, CompletePaths: 0, ExecutionTimeMs: execTimeMs}
	}
	return BFSResult{Tree: tree, NodeCount: nodeCount, CompletePaths: 1, ExecutionTimeMs: execTimeMs}
}

func bfsShortest(target *Graph.ElementGraph, root *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph, nodeCount *int, found *bool) {
	queue := NewQueue()
	queue.Enqueue(&QueueItem{GraphNode: target, TreeNode: root})

	for !queue.IsEmpty() && !*found {
		item := queue.Dequeue()
		current := item.GraphNode
		node := item.TreeNode

		*nodeCount++

		if Graph.IsLeaf(current, graphMap) {
			*found = true
			return
		}

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

			queue.Enqueue(&QueueItem{GraphNode: recipe.FirstElement, TreeNode: left})
			queue.Enqueue(&QueueItem{GraphNode: recipe.SecondElement, TreeNode: right})

			if Graph.IsLeaf(recipe.FirstElement, graphMap) && Graph.IsLeaf(recipe.SecondElement, graphMap) {
				*found = true
				return
			}
		}
	}
}
