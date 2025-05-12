package bfs

import (
	"runtime"
	"sync"
	"time"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

// Output
type BFSResult struct {
	Tree            *Tree.Tree `json:"tree"`
	NodeCount       int        `json:"nodeCount"`
	CompletePaths   int        `json:"completePaths"`
	ExecutionTimeMs int64      `json:"executionTimeMs"`
	Done            bool       `json:"done"`
}

// =====================================BFS Multiple Recipe=======================================

func FindMultiplePath(
	target *Graph.ElementGraph,
	maxPaths int,
	delay time.Duration,
	updates chan<- BFSResult,
	_ map[string]*Graph.ElementGraph,
) BFSResult {
	start := time.Now()

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
		nodeCount int
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
			mu.Lock()
			nodeCount++
			mu.Unlock()

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

				if recipe.FirstElement.Tier > parentTier || recipe.SecondElement.Tier > parentTier {
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

					time.Sleep(delay)

					mu.Lock()
					parent.Children = append(parent.Children, step)

					if checkFullPath(&step) {
						pathCount++
					}

					left.Parent = &step
					right.Parent = &step

					if updates != nil {
						cloned := Tree.CopyTree(tree)
						updates <- BFSResult{
							Done:            false,
							Tree:            cloned,
							NodeCount:       nodeCount,
							CompletePaths:   pathCount,
							ExecutionTimeMs: time.Since(start).Milliseconds(),
						}
					}

					if pathCount == maxPaths {
						done = true
						mu.Unlock()
						return
					}
					mu.Unlock()

					if !Graph.IsLeafNode(rc.FirstElement) {
						next = append(next, bfsItem{rc.FirstElement, left})
					}
					if !Graph.IsLeafNode(rc.SecondElement) {
						next = append(next, bfsItem{rc.SecondElement, right})
					}

				}(recipe, itm.node)
			}
		}

		wg.Wait()
		queue = next
	}

	Tree.PruneTreeToBasicPaths(tree.First)
	return BFSResult{
		Tree:            tree,
		NodeCount:       nodeCount,
		CompletePaths:   pathCount,
		ExecutionTimeMs: time.Since(start).Milliseconds(),
		Done:            true,
	}
}

// =====================================BFS First Recipe=======================================

func FindFirstPath(target *Graph.ElementGraph, delay time.Duration, updates chan<- BFSResult) BFSResult {
	start := time.Now()

	root := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: root}
	nodeCount := 0
	found := false

	pureBfS(target, root, &nodeCount, &found, delay, updates, tree)
	Tree.PruneTreeToBasicPaths(root)

	execTimeMs := time.Since(start).Milliseconds()
	if !found {
		return BFSResult{
			Tree:            tree,
			NodeCount:       nodeCount,
			CompletePaths:   0,
			ExecutionTimeMs: execTimeMs,
			Done:            true,
		}
	}
	return BFSResult{
		Tree:            tree,
		NodeCount:       nodeCount,
		CompletePaths:   1,
		ExecutionTimeMs: execTimeMs,
		Done:            true,
	}
}

func pureBfS(target *Graph.ElementGraph, root *Tree.TreeNodeElement, nodeCount *int, found *bool, delay time.Duration, updates chan<- BFSResult, tree *Tree.Tree) {
	type bfsItem struct {
		elem *Graph.ElementGraph
		node *Tree.TreeNodeElement
	}

	queue := []bfsItem{{target, root}}

	for len(queue) > 0 && !*found {
		var next []bfsItem

		for _, itm := range queue {
			(*nodeCount)++

			if len(itm.elem.Recipes) == 0 {
				continue
			}

			parentTier := itm.elem.Tier

			for _, recipe := range itm.elem.Recipes {
				if recipe.FirstElement.Tier > parentTier || recipe.SecondElement.Tier > parentTier {
					continue
				}

				left := &Tree.TreeNodeElement{Name: recipe.FirstElement.Name}
				right := &Tree.TreeNodeElement{Name: recipe.SecondElement.Name}
				recipeNode := Tree.TreeNodeRecipe{
					FirstElement:  left,
					SecondElement: right,
					ResultElement: itm.node,
				}
				left.Parent = &recipeNode
				right.Parent = &recipeNode

				itm.node.Children = append(itm.node.Children, recipeNode)

				if delay > 0 {
					time.Sleep(delay)
				}

				if updates != nil {
					cloned := Tree.CopyTree(tree)
					updates <- BFSResult{
						Tree:            cloned,
						NodeCount:       *nodeCount,
						CompletePaths:   0,
						ExecutionTimeMs: time.Since(time.Now().Add(-delay)).Milliseconds(),
						Done:            false,
					}
				}

				if checkFullPath(&recipeNode) {
					*found = true
					return
				}

				if !Graph.IsLeafNode(recipe.FirstElement) {
					next = append(next, bfsItem{recipe.FirstElement, left})
				}
				if !Graph.IsLeafNode(recipe.SecondElement) {
					next = append(next, bfsItem{recipe.SecondElement, right})
				}
			}
		}
		queue = next
	}
}

// =====================================Helper=======================================

func checkFullPath(r *Tree.TreeNodeRecipe) bool {
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

func isRecipeSolved(r *Tree.TreeNodeRecipe) bool {
	return isSolvedElem(r.FirstElement) && isSolvedElem(r.SecondElement)
}

func isSolvedElem(n *Tree.TreeNodeElement) bool {
	if len(n.Children) == 0 {
		return Graph.IsLeafName(n.Name)
	}
	for _, c := range n.Children {
		if isRecipeSolved(&c) {
			return true
		}
	}
	return false
}
