package dfs

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

// DFSResult is the output type for DFS algorithms.
type DFSResult struct {
	Tree            *Tree.Tree `json:"tree"`
	NodeCount       int        `json:"nodeCount"`
	CompletePaths   int        `json:"completePaths"`
	ExecutionTimeMs int64      `json:"executionTimeMs"`
	Done            bool       `json:"done"`
}

func FindMultiplePathDree(
	target *Graph.ElementGraph,
	maxPaths int,
	delay time.Duration,
	updates chan<- DFSResult,
	graphMap map[string]*Graph.ElementGraph,
) DFSResult {
	start := time.Now()

	rootDree := &Tree.TreeNodeElement{Name: target.Name}
	tree := &Tree.Tree{First: rootDree}
	var nodeCount int32 = 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	_ = dfsMultiThreadingDree(rootDree, target, graphMap, maxPaths, updates, &nodeCount, delay, &mu, &wg, start, tree)
	wg.Wait()

	return DFSResult{
		Tree:            tree,
		NodeCount:       int(atomic.LoadInt32(&nodeCount)),
		CompletePaths:   tree.First.BranchCount,
		ExecutionTimeMs: time.Since(start).Milliseconds(),
		Done:            true,
	}
}

var basicCache sync.Map

func dfsMultiThreadingDree(
	node *Tree.TreeNodeElement,
	elem *Graph.ElementGraph,
	graphMap map[string]*Graph.ElementGraph,
	maxPaths int,
	updates chan<- DFSResult,
	nodeCountPtr *int32,
	delay time.Duration,
	mu *sync.Mutex,
	wg *sync.WaitGroup,
	start time.Time,
	tree *Tree.Tree,
) int {
	atomic.AddInt32(nodeCountPtr, 1)

	if Tree.IsBasic(elem.Name) || len(elem.Recipes) == 0 {
		if val, ok := basicCache.Load(elem.Name); ok {
			shared := val.(*Tree.TreeNodeElement)
			*node = *shared
		} else {
			node.BranchCount = 1
			node.Name = elem.Name
			basicCache.Store(elem.Name, &Tree.TreeNodeElement{
				Name:        elem.Name,
				BranchCount: 1,
			})
		}
		return 1
	}

	node.Children = []Tree.TreeNodeRecipe{}

	var total int64
	var localWg sync.WaitGroup

	for _, rc := range elem.Recipes {
		if rc.FirstElement.Tier >= elem.Tier || rc.SecondElement.Tier >= elem.Tier {
			continue
		}

		wg.Add(1)
		localWg.Add(1)

		go func(r Graph.Recipe) {
			defer wg.Done()
			defer localWg.Done()

			left := &Tree.TreeNodeElement{Name: r.FirstElement.Name}
			right := &Tree.TreeNodeElement{Name: r.SecondElement.Name}

			leftCnt := dfsMultiThreadingDree(left, r.FirstElement, graphMap, maxPaths, updates, nodeCountPtr, delay, mu, wg, start, tree)
			rightCnt := dfsMultiThreadingDree(right, r.SecondElement, graphMap, maxPaths, updates, nodeCountPtr, delay, mu, wg, start, tree)

			mu.Lock()
			node.Children = append(node.Children, Tree.TreeNodeRecipe{
				FirstElement:  left,
				SecondElement: right,
				ResultElement: node,
			})
			mu.Unlock()

			left.BranchCount = leftCnt
			right.BranchCount = rightCnt

			prod := leftCnt * rightCnt
			atomic.AddInt64(&total, int64(prod))

		}(rc)
	}

	localWg.Wait()

	node.BranchCount = int(total)
	if node.BranchCount > maxPaths {
		for node.BranchCount > maxPaths {
			node.BranchCount = cutChildren(node, node.BranchCount-maxPaths)
		}
	}

	if delay > 0 {
		time.Sleep(delay)
	}

	if updates != nil {
		snapshot := func() *Tree.Tree {
			mu.Lock()
			defer mu.Unlock()
			return &Tree.Tree{
				First: CopySubtree(node),
			}
		}()

		updates <- DFSResult{
			Tree:            snapshot,
			NodeCount:       int(atomic.LoadInt32(nodeCountPtr)),
			CompletePaths:   node.BranchCount,
			ExecutionTimeMs: time.Since(start).Milliseconds(),
			Done:            false,
		}
	}

	return node.BranchCount
}

func CopySubtree(node *Tree.TreeNodeElement) *Tree.TreeNodeElement {
	if node == nil {
		return nil
	}
	clone := &Tree.TreeNodeElement{
		Name:        node.Name,
		BranchCount: node.BranchCount,
	}
	for _, rc := range node.Children {
		newLeft := CopySubtree(rc.FirstElement)
		newRight := CopySubtree(rc.SecondElement)
		clone.Children = append(clone.Children, Tree.TreeNodeRecipe{
			FirstElement:  newLeft,
			SecondElement: newRight,
			ResultElement: clone,
		})
	}
	return clone
}

func cutChildren(node *Tree.TreeNodeElement, totalCut int) int {
	if totalCut <= 0 || len(node.Children) == 0 {
		return node.BranchCount
	}

	type prodInfo struct {
		idx int
		val int
	}
	infos := make([]prodInfo, len(node.Children))
	for i, rc := range node.Children {
		if rc.FirstElement == nil || rc.SecondElement == nil {
			continue
		}
		infos[i] = prodInfo{
			idx: i,
			val: rc.FirstElement.BranchCount * rc.SecondElement.BranchCount,
		}
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

	dfsShortest(target, root, graphMap, &nodeCount, delay, updates, tree, start) //  &mu

	return DFSResult{
		Tree:            tree,
		NodeCount:       nodeCount,
		CompletePaths:   1,
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
