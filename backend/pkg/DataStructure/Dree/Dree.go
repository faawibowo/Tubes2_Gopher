package Dree

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
)

type Dree struct {
	First *DreeNodeElement `json:"first"`
}

type DreeNodeElement struct {
	Name        string           `json:"name"`
	Children    []DreeNodeRecipe `json:"children"`
	Parent      *DreeNodeRecipe  `json:"-"` // to avoid circular reference in JSON
	BranchCount int              `json:"-"`
}

type DreeNodeRecipe struct {
	FirstElement  *DreeNodeElement `json:"firstElement"`
	SecondElement *DreeNodeElement `json:"secondElement"`
	ResultElement *DreeNodeElement `json:"-"` // removed from JSON
}

var basicCache sync.Map

func BuildDreeDFSConcurrent(
	target *Graph.ElementGraph,
	maxPaths int,
	graphMap map[string]*Graph.ElementGraph,
) *Dree {
	basicCache = sync.Map{}

	root := &DreeNodeElement{Name: target.Name}

	_ = dfs(root, target, graphMap, maxPaths)

	return &Dree{First: root}
}

func dfs(
	node *DreeNodeElement,
	elem *Graph.ElementGraph,
	graphMap map[string]*Graph.ElementGraph,
	maxPaths int,
) int {
	if isBasicDree(elem) || len(elem.Recipes) == 0 {
		if val, ok := basicCache.Load(elem.Name); ok {
			shared := val.(*DreeNodeElement)
			*node = *shared
		} else {
			node.BranchCount = 1
			basicCache.Store(elem.Name, &DreeNodeElement{
				Name:        elem.Name,
				BranchCount: 1,
			})
		}
		return 1
	}

	node.Children = []DreeNodeRecipe{}

	var wg sync.WaitGroup
	var total int64
	var mu sync.Mutex

	for _, rc := range elem.Recipes {
		if rc.FirstElement.Tier >= elem.Tier || rc.SecondElement.Tier >= elem.Tier {
			continue
		}
		wg.Add(1)
		go func(r Graph.Recipe) {
			defer wg.Done()

			left := &DreeNodeElement{Name: r.FirstElement.Name}
			right := &DreeNodeElement{Name: r.SecondElement.Name}

			leftCnt := dfs(left, r.FirstElement, graphMap, maxPaths)
			rightCnt := dfs(right, r.SecondElement, graphMap, maxPaths)

			mu.Lock()
			node.Children = append(node.Children, DreeNodeRecipe{
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

	wg.Wait()
	node.BranchCount = int(total)

	if node.BranchCount > maxPaths {
		for node.BranchCount > maxPaths {
			node.BranchCount = cutChildren(node, node.BranchCount-maxPaths)
		}
	}

	return node.BranchCount
}

func cutChildren(node *DreeNodeElement, totalCut int) int {
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

			*rc = DreeNodeRecipe{} // overwrite to blank
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
		node.Children[i] = DreeNodeRecipe{}
	}
	node.Children = valid

	node.BranchCount = currTotal
	return currTotal
}

func isBasicDree(e *Graph.ElementGraph) bool {
	return e.Name == "Air" || e.Name == "Water" || e.Name == "Earth" || e.Name == "Fire"
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func PrintDreeTree(node *DreeNodeElement, depth int, printName bool, prefixStack []bool) {
	if node == nil {
		return
	}
	prefix := ""
	for _, active := range prefixStack {
		if active {
			prefix += "│  "
		} else {
			prefix += "   "
		}
	}

	if printName {
		fmt.Println(prefix + node.Name)
	}

	for i, recipe := range node.Children {
		isLast := i == len(node.Children)-1

		recipePrefix := prefix + "├─ Recipe:"
		if isLast {
			recipePrefix = prefix + "└─ Recipe:"
		}
		fmt.Println(recipePrefix)

		childStack := append(prefixStack, !isLast)

		leftPrefix := ""
		for _, active := range childStack {
			if active {
				leftPrefix += "│  "
			} else {
				leftPrefix += "   "
			}
		}
		fmt.Println(leftPrefix + "├─ " + recipe.FirstElement.Name + " (left)")
		PrintDreeTree(recipe.FirstElement, depth+1, false, append(childStack, true))

		rightPrefix := ""
		for _, active := range childStack {
			if active {
				rightPrefix += "│  "
			} else {
				rightPrefix += "   "
			}
		}
		fmt.Println(rightPrefix + "└─ " + recipe.SecondElement.Name + " (right)")
		PrintDreeTree(recipe.SecondElement, depth+1, false, append(childStack, false))
	}
}
