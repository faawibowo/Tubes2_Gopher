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

			// recipeNode :=
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

			// if updates != nil {
			// 	cloned := Tree.CopyTree(tree)
			// 	updates <- DFSResult{
			// 		Tree:            cloned,
			// 		NodeCount:       *nodeCountPtr,
			// 		CompletePaths:   node.BranchCount,
			// 		ExecutionTimeMs: time.Since(start).Milliseconds(),
			// 		Done:            false,
			// 	}
			// }
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
				First: CopySubtree(node), // 👈 only copy the subtree under `node`
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
			ResultElement: clone, // 👈 point back to itself
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

//// page break//////
// basis
// kalau leaf branch countnya = 1
// else
//
// bikin node baru (tree dapet dari parameter pointer)
// dapet list of recipes dari graphMap
// iterate through all the recipes, satu2 construct node baru dengan dfs, disimpan dalam children
// seluruh children simpen informasi branch count
// branch count kita dihitung dari sigma recipe (firstElement.branchCount * secondElement.branchCount)
// validation :
// kalau current branch count lebih dari max path, cut branch yang ada di children. (fungsi baru)
// new Branch count hasil cut ditambah ke pathCounts (pointer dari parameter)
// return node ? or dfs result

// fungsi baru yang ngecut children, parameter: node, jumlah cut
// dari node, masuk ke recipes
// sebagaimana mungkin kita mau cut biar jumlah first*second di recipes <= new branch count
// misalnya recipes jumlah first*second = (3,4,5,2,2), dan jumlah cutnya = 5
// recipes baru jadi = (3,4,4,0,0),
// yang bisa langsung dibuang lgsg di buang, jadi
// recipes baru (3,4,4)
// yang recipes kena berubah kita liat first.branchCount dan second.branchCountnya berapa
// misalnya first * second = 5 * 1, maka kita mau buat biar hasil perkaliannya <= 4
// jadi first * second yang baru adalah 4 * 1.
// oleh karena itu kita rekursiin dengnan fungsi ini dengan parameter (first, 1 ), (1 karena itu jumlah yang mau di cut dari first)
// kalau first * second misalnya 7 * 4, dan ingin di cut hingga first * second maks 10, maka first * second baru nya adalah 5*2
// kalau first * second misalnya 4 * 4 dan ingin di cut biar maks perkaliannya 10, ya paling first * second jadi 3 * 3 yang penting kurang dari maks

// func BuildRecipeTree(
// 	target *Graph.ElementGraph,
// 	graphMap map[string]*Graph.ElementGraph,
// 	maxCount int,
// 	delay time.Duration,
// 	updates chan<- DFSResult,
// ) DFSResult {
// 	start := time.Now()

// 	root := &Tree.TreeNodeElement{Name: target.Name}
// 	tree := &Tree.Tree{First: root}

// 	var nodeCount int = 0
// 	var pathCount int = 0
// 	var mu sync.Mutex
// 	var wg sync.WaitGroup

// 	dfsMultiThreading(target, root, graphMap, &nodeCount, &pathCount, maxCount, delay, updates, &mu, &wg, tree, start)
// 	wg.Wait()

// 	pruneIncompletePaths(root, graphMap)
// 	wg.Wait()

// 	return DFSResult{
// 		Tree:            tree,
// 		NodeCount:       nodeCount,
// 		CompletePaths:   pathCount,
// 		ExecutionTimeMs: time.Since(start).Milliseconds(),
// 		Done:            true,
// 	}
// }

// func isEqual(a int, b int) bool {
// 	return a == b
// }

// func isTreeLeaf(left *Tree.TreeNodeElement, right *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph) bool {
// 	return Graph.IsLeafNode(graphMap[left.Name]) && Graph.IsLeafNode(graphMap[right.Name])
// }

// func dfsMultiThreading(
// 	current *Graph.ElementGraph,
// 	node *Tree.TreeNodeElement,
// 	graphMap map[string]*Graph.ElementGraph,
// 	nodeCountPtr *int,
// 	pathCountPtr *int,
// 	maxCount int,
// 	delay time.Duration,
// 	updates chan<- DFSResult,
// 	mu *sync.Mutex,
// 	wg *sync.WaitGroup,
// 	tree *Tree.Tree,
// 	start time.Time,
// ) {
// 	if isEqual(*pathCountPtr, maxCount) {
// 		return
// 	}

// 	mu.Lock()
// 	(*nodeCountPtr)++
// 	mu.Unlock()

// 	for _, recipe := range current.Recipes {
// 		if isEqual(*pathCountPtr, maxCount) {
// 			return
// 		}

// 		wg.Add(1)
// 		time.Sleep(delay)
// 		go func(r Graph.Recipe) {
// 			defer wg.Done()

// 			left := &Tree.TreeNodeElement{Name: r.FirstElement.Name}
// 			right := &Tree.TreeNodeElement{Name: r.SecondElement.Name}
// 			recipeNode := Tree.TreeNodeRecipe{
// 				FirstElement:  left,
// 				SecondElement: right,
// 				ResultElement: node,
// 			}

// 			mu.Lock()
// 			if isEqual(*pathCountPtr, maxCount) {
// 				mu.Unlock()
// 				return
// 			}
// 			node.Children = append(node.Children, recipeNode)
// 			if isTreeLeaf(left, right, graphMap) {
// 				(*pathCountPtr)++
// 			}
// 			mu.Unlock()

// 			left.Parent = &recipeNode
// 			right.Parent = &recipeNode

// 			if updates != nil {
// 				cloned := Tree.CopyTree(tree)
// 				updates <- DFSResult{
// 					Tree:            cloned,
// 					NodeCount:       *nodeCountPtr,
// 					CompletePaths:   *pathCountPtr,
// 					ExecutionTimeMs: time.Since(start).Milliseconds(),
// 					Done:            false,
// 				}
// 			}

// 			if r.FirstElement.Tier >= current.Tier || r.SecondElement.Tier >= current.Tier {
// 				return
// 			}

// 			if isEqual(*pathCountPtr, maxCount) {
// 				return
// 			}
// 			dfsMultiThreading(r.FirstElement, left, graphMap, nodeCountPtr, pathCountPtr, maxCount, delay, updates, mu, wg, tree, start)
// 			dfsMultiThreading(r.SecondElement, right, graphMap, nodeCountPtr, pathCountPtr, maxCount, delay, updates, mu, wg, tree, start)
// 		}(recipe)
// 	}
// }

// func pruneIncompletePaths(node *Tree.TreeNodeElement, graphMap map[string]*Graph.ElementGraph) {

// 	if node == nil {
// 		return
// 	}

// 	validRecipes := make([]Tree.TreeNodeRecipe, 0, len(node.Children))

// 	for _, recipe := range node.Children {
// 		pruneIncompletePaths(recipe.FirstElement, graphMap)
// 		pruneIncompletePaths(recipe.SecondElement, graphMap)

// 		if len(recipe.FirstElement.Children) == 0 && len(recipe.SecondElement.Children) == 0 {
// 			leaf1, leaf2 := false, false
// 			if elem, ok := graphMap[recipe.FirstElement.Name]; ok {
// 				leaf1 = Graph.IsLeafNode(elem)
// 			}
// 			if elem, ok := graphMap[recipe.SecondElement.Name]; ok {
// 				leaf2 = Graph.IsLeafNode(elem)
// 			}

// 			if leaf1 && leaf2 {
// 				validRecipes = append(validRecipes, recipe)
// 			}
// 		} else {
// 			validRecipes = append(validRecipes, recipe)
// 		}
// 	}
// 	node.Children = validRecipes
// }

// =====================================DFS First Recipe=======================================

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

// usable -> BuildRecipeTree(), FindShortestPath()
