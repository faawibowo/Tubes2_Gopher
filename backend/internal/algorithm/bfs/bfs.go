package bfs

import (
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
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
	root.BranchCount = 0
	tree := &Tree.Tree{First: root}

	type bfsItem struct {
		elem *Graph.ElementGraph
		node *Tree.TreeNodeElement
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		nodeCount int32
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

	queue := []bfsItem{{elem: target, node: root}}

	for len(queue) > 0 && !done {
		var next []bfsItem
		var nextMu sync.Mutex

		for _, itm := range queue {
			atomic.AddInt32(&nodeCount, 1)

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
					}
					left.Parent = &step
					right.Parent = &step

					if updates != nil {
						cloned := Tree.CopyTree(tree)
						updates <- BFSResult{
							Done:            false,
							Tree:            cloned,
							NodeCount:       int(atomic.LoadInt32(&nodeCount)),
							CompletePaths:   0,
							ExecutionTimeMs: time.Since(start).Milliseconds(),
						}
					}
					if false {
						done = true
						mu.Unlock()
						return
					}
					mu.Unlock()

					if !Graph.IsLeafNode(rc.FirstElement) {
						nextMu.Lock()
						next = append(next, bfsItem{elem: rc.FirstElement, node: left})
						nextMu.Unlock()
					}
					if !Graph.IsLeafNode(rc.SecondElement) {
						nextMu.Lock()
						next = append(next, bfsItem{elem: rc.SecondElement, node: right})
						nextMu.Unlock()
					}
				}(recipe, itm.node)
			}
		}
		wg.Wait()
		queue = next
	}

	completePaths := recalcBranchCount(root)
	for root.BranchCount > maxPaths {
		root.BranchCount = cutChildren(root, root.BranchCount-maxPaths)
		completePaths = recalcBranchCount(root)
	}

	return BFSResult{
		Tree:            tree,
		NodeCount:       int(atomic.LoadInt32(&nodeCount)),
		CompletePaths:   completePaths,
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
	prunePath(root)

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
			break
		}
		curr = curr.ResultElement.Parent
	}
	return true
}

func isRecipeSolved(r *Tree.TreeNodeRecipe) bool {
	return isSolvedElem(r.FirstElement) && isSolvedElem(r.SecondElement)
}

func isSolvedElem(n *Tree.TreeNodeElement) bool {

	if n == nil {
		return false
	}

	if Graph.IsLeafName((n.Name)) {
		return true
	}

	for _, c := range n.Children {
		if isRecipeSolved(&c) {
			return true
		}
	}
	return false
}

func prunePath(r *Tree.TreeNodeElement) {
	if r == nil {
		return
	}

	var queue []*Tree.TreeNodeElement
	queue = append(queue, r)

	for len(queue) > 0 {
		nodeElement := queue[0]
		queue = queue[1:]

		if len(nodeElement.Children) > 0 {
			nodeElement.Children = nodeElement.Children[:1]
			if nodeElement.Children[0].FirstElement != nil {
				queue = append(queue, nodeElement.Children[0].FirstElement)
			}
			if nodeElement.Children[0].SecondElement != nil {
				queue = append(queue, nodeElement.Children[0].SecondElement)
			}
		}
	}
}

func recalcBranchCount(n *Tree.TreeNodeElement) int {
	if n == nil {
		return 0
	}
	if len(n.Children) == 0 {
		n.BranchCount = 1
		return 1
	}
	total := 0
	for i := range n.Children {
		r := &n.Children[i]
		left := recalcBranchCount(r.FirstElement)
		right := recalcBranchCount(r.SecondElement)
		total += left * right
	}
	n.BranchCount = total
	return total
}

func cutChildren(node *Tree.TreeNodeElement, totalCut int) int {
	if totalCut <= 0 || len(node.Children) == 0 {
		return node.BranchCount
	}

	type prodInfo struct {
		idx int
		val int
	}
	infos := make([]prodInfo, 0, len(node.Children))
	for i, rc := range node.Children {
		// Only consider valid children.
		if rc.FirstElement == nil || rc.SecondElement == nil {
			continue
		}
		p := rc.FirstElement.BranchCount * rc.SecondElement.BranchCount
		infos = append(infos, prodInfo{idx: i, val: p})
	}

	sort.Slice(infos, func(i, j int) bool { return infos[i].val > infos[j].val })

	currTotal := node.BranchCount

	for _, info := range infos {
		if totalCut <= 0 {
			break
		}
		rc := &node.Children[info.idx]
		if info.val <= totalCut {
			totalCut -= info.val
			currTotal -= info.val
			rc.FirstElement = nil
			rc.SecondElement = nil
			rc.ResultElement = nil
			*rc = Tree.TreeNodeRecipe{}
			continue
		}

		need := totalCut
		if rc.FirstElement.BranchCount >= rc.SecondElement.BranchCount {
			maxDel := rc.FirstElement.BranchCount - 1
			wantDel := min(maxDel, (need+rc.SecondElement.BranchCount-1)/rc.SecondElement.BranchCount)
			cutChildren(rc.FirstElement, wantDel)
		} else {
			maxDel := rc.SecondElement.BranchCount - 1
			wantDel := min(maxDel, (need+rc.FirstElement.BranchCount-1)/rc.FirstElement.BranchCount)
			cutChildren(rc.SecondElement, wantDel)
		}

		newProd := rc.FirstElement.BranchCount * rc.SecondElement.BranchCount
		diff := info.val - newProd
		totalCut -= diff
		currTotal -= diff
	}

	valid := node.Children[:0]
	for _, c := range node.Children {
		if c.FirstElement != nil && c.SecondElement != nil &&
			c.FirstElement.BranchCount > 0 && c.SecondElement.BranchCount > 0 {
			valid = append(valid, c)
		}
	}
	for i := len(valid); i < len(node.Children); i++ {
		node.Children[i] = Tree.TreeNodeRecipe{}
	}
	node.Children = valid

	node.BranchCount = currTotal
	return currTotal
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
