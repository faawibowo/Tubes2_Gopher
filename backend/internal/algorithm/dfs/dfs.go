package dfs

import (
	"sort"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

func FindAllPaths(node *Graph.Graph, graphMap map[string]*Graph.Graph) []*Tree.Tree {
	var paths []*Tree.Tree
	if node == nil {
		return paths
	}
	rootTree := Tree.NewTree(node)
	dfs(node, rootTree, graphMap, &paths)

	// Sort
	sort.Slice(paths, func(i, j int) bool {
		return Tree.CountNodes(paths[i]) < Tree.CountNodes(paths[j])
	})
	return paths
}

func dfs(node *Graph.Graph, currentTree *Tree.Tree, graphMap map[string]*Graph.Graph, result *[]*Tree.Tree) {
	// Basis
	if node == nil {
		return
	}

	// If leaf -> append  deep copy
	if Graph.IsLeaf(node, graphMap) {
		*result = append(*result, Tree.CopyTree(currentTree))
		return
	}

	// Iterate semua combination
	for _, combi := range node.Combinations {
		branchTree := Tree.CopyTree(currentTree)
		branchTree.FirstChild = Tree.NewTree(combi.FirstElement)
		branchTree.SecondChild = Tree.NewTree(combi.SecondElement)

		dfs(combi.FirstElement, branchTree.FirstChild, graphMap, result)
		dfs(combi.SecondElement, branchTree.SecondChild, graphMap, result)
	}
}

func FindShortestPath(trees []*Tree.Tree) *Tree.Tree {
	if len(trees) == 0 {
		return nil
	}
	return trees[0]
}
