package dfs

import (
	"sort"
	"sync"
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

// Dree and its nodes represent the DFS tree structure.
type Dree struct {
	First *DreeNodeElement `json:"first"`
}

type DreeNodeElement struct {
	Name        string           `json:"name"`
	Children    []DreeNodeRecipe `json:"children"`
	Parent      *DreeNodeRecipe  `json:"-"` // do not include in JSON to avoid circular refs
	BranchCount int
}

type DreeNodeRecipe struct {
	FirstElement  *DreeNodeElement `json:"firstElement"`
	SecondElement *DreeNodeElement `json:"secondElement"`
	ResultElement *DreeNodeElement `json:"-"` // removed from JSON
}

func FindMultiplePathDree(
	target *Graph.ElementGraph,
	maxPaths int,
	graphMap map[string]*Graph.ElementGraph,
	delay time.Duration,
) *DFSResult {
	start := time.Now()

	// Build the internal Dree tree.
	rootDree := &DreeNodeElement{Name: target.Name, BranchCount: 1}
	var nodeCount int = 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	dfsMultiThreadingDree(target, rootDree, graphMap, &nodeCount, maxPaths, delay, &mu, &wg)
	wg.Wait()

	// Apply branch cutting if the branch count exceeds maxPaths.
	if rootDree.BranchCount > maxPaths {
		cutChildren(rootDree, rootDree.BranchCount-maxPaths)
	}

	// Convert the Dree tree into a standard Tree.
	tree := convertDreeToTree(&Dree{First: rootDree})

	return &DFSResult{
		Tree:            tree,
		NodeCount:       nodeCount,
		CompletePaths:   1, // For DFS, if at least one path is found.
		ExecutionTimeMs: time.Since(start).Milliseconds(),
		Done:            true,
	}
}

// dfsMultiThreadingDree concurrently builds the internal Dree tree.
// It spawns a goroutine for each recipe, updates nodeCount, and sets BranchCount
// as the product of its children’s branch counts. (For leaf nodes, BranchCount remains 1.)
func dfsMultiThreadingDree(
	current *Graph.ElementGraph,
	node *DreeNodeElement,
	graphMap map[string]*Graph.ElementGraph,
	nodeCountPtr *int,
	maxPaths int,
	delay time.Duration,
	mu *sync.Mutex,
	wg *sync.WaitGroup,
) {
	mu.Lock()
	*nodeCountPtr++
	mu.Unlock()

	// If there are no recipes, the node is a leaf.
	if len(current.Recipes) == 0 || Graph.IsLeafNode(current) {
		node.BranchCount = 1
		return
	}

	// For each recipe, spawn concurrent routines.
	for _, recipe := range current.Recipes {
		wg.Add(1)
		time.Sleep(delay)
		go func(r Graph.Recipe) {
			defer wg.Done()

			// Create left and right Dree nodes.
			left := &DreeNodeElement{Name: r.FirstElement.Name, BranchCount: 1}
			right := &DreeNodeElement{Name: r.SecondElement.Name, BranchCount: 1}
			recipeNode := DreeNodeRecipe{
				FirstElement:  left,
				SecondElement: right,
				ResultElement: node,
			}

			// Append the recipe (thread-safe).
			mu.Lock()
			node.Children = append(node.Children, recipeNode)
			mu.Unlock()

			// Set parent pointers.
			left.Parent = &recipeNode
			right.Parent = &recipeNode

			// Recurse concurrently into subtrees.
			dfsMultiThreadingDree(r.FirstElement, left, graphMap, nodeCountPtr, maxPaths, delay, mu, wg)
			dfsMultiThreadingDree(r.SecondElement, right, graphMap, nodeCountPtr, maxPaths, delay, mu, wg)

			// Update the BranchCount for this recipe.
			branchProd := left.BranchCount * right.BranchCount

			// Safely update the current node's BranchCount.
			mu.Lock()
			node.BranchCount += branchProd
			mu.Unlock()
		}(recipe)
	}
}

// ----------------------------------------------------------------------------
// CUT CHILDREN & CONVERSION (UNCHANGED)
// ----------------------------------------------------------------------------

// cutChildren reduces children branch counts so that node.BranchCount becomes ≤ originalTotal - sisaToCut.
// sisaToCut is the amount that needs to be eliminated.
func cutChildren(node *DreeNodeElement, sisaToCut int) int {
	if node == nil || sisaToCut <= 0 {
		return node.BranchCount
	}

	// prodInfo holds the product and index for each child recipe.
	type prodInfo struct {
		idx int
		val int
	}
	var infos []prodInfo

	// Calculate product for each recipe.
	for i, rc := range node.Children {
		prod := rc.FirstElement.BranchCount * rc.SecondElement.BranchCount
		infos = append(infos, prodInfo{idx: i, val: prod})
	}

	// Sort recipes from highest product to lowest.
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].val > infos[j].val
	})

	total := node.BranchCount

	// Iterate over sorted recipes, reducing branch counts until sisaToCut is removed.
	for _, info := range infos {
		if sisaToCut <= 0 {
			break
		}
		rc := &node.Children[info.idx]
		oldProd := rc.FirstElement.BranchCount * rc.SecondElement.BranchCount

		// Remove whole recipe if its product is less than or equal to needed cut.
		if oldProd <= sisaToCut {
			sisaToCut -= oldProd
			total -= oldProd
			rc.FirstElement = nil
			rc.SecondElement = nil
			continue
		}

		// Partial cut: reduce the side with the larger branch count.
		need := sisaToCut
		if rc.FirstElement.BranchCount >= rc.SecondElement.BranchCount {
			needLeft := min(rc.FirstElement.BranchCount-1, (need+rc.SecondElement.BranchCount-1)/rc.SecondElement.BranchCount)
			newLeft := cutChildren(rc.FirstElement, needLeft)
			rc.FirstElement.BranchCount = newLeft
		} else {
			needRight := min(rc.SecondElement.BranchCount-1, (need+rc.FirstElement.BranchCount-1)/rc.FirstElement.BranchCount)
			newRight := cutChildren(rc.SecondElement, needRight)
			rc.SecondElement.BranchCount = newRight
		}

		newProd := rc.FirstElement.BranchCount * rc.SecondElement.BranchCount
		diff := oldProd - newProd
		sisaToCut -= diff
		total -= diff
	}

	// Remove completely cut (deleted) recipes.
	var pruned []DreeNodeRecipe
	for _, rc := range node.Children {
		if rc.FirstElement != nil && rc.SecondElement != nil {
			pruned = append(pruned, rc)
		}
	}
	node.Children = pruned

	// Recalculate node's branch count.
	if len(node.Children) == 0 {
		node.BranchCount = 1
	} else {
		subTotal := 0
		for _, rc := range node.Children {
			subTotal += rc.FirstElement.BranchCount * rc.SecondElement.BranchCount
		}
		node.BranchCount = subTotal
	}

	return node.BranchCount
}

// min returns the smaller of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// convertDreeToTree converts the internal Dree structure into a standard Tree.
func convertDreeToTree(d *Dree) *Tree.Tree {
	if d == nil || d.First == nil {
		return nil
	}
	return &Tree.Tree{First: convertDreeNode(d.First)}
}

func convertDreeNode(dn *DreeNodeElement) *Tree.TreeNodeElement {
	if dn == nil {
		return nil
	}
	newNode := &Tree.TreeNodeElement{Name: dn.Name}
	for _, dr := range dn.Children {
		recipe := Tree.TreeNodeRecipe{
			FirstElement:  convertDreeNode(dr.FirstElement),
			SecondElement: convertDreeNode(dr.SecondElement),
			ResultElement: newNode,
		}
		if recipe.FirstElement != nil {
			recipe.FirstElement.Parent = &recipe
		}
		if recipe.SecondElement != nil {
			recipe.SecondElement.Parent = &recipe
		}
		newNode.Children = append(newNode.Children, recipe)
	}
	return newNode
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
