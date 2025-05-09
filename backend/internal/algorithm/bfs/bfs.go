package bfs

import (
	"sort"

	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Graph"
	"github.com/faawibowo/Tubes2_Gopher/pkg/DataStructure/Tree"
)

type branch struct {
	graph *Graph.Graph
	tree  *Tree.Tree
}

func FindAllPaths(node *Graph.Graph, graphMap map[string]*Graph.Graph) []*Tree.Tree {
	paths := bfs(node, graphMap)

	// Sort
	sort.Slice(paths, func(i, j int) bool {
		return Tree.CountNodes(paths[i]) < Tree.CountNodes(paths[j])
	})
	return paths
}

func bfs(node *Graph.Graph, graphMap map[string]*Graph.Graph) []*Tree.Tree {
	var result []*Tree.Tree
	if node == nil {
		return result
	}

	// Initialize branch
	rootTree := Tree.NewTree(node)
	var workQueue []branch
	workQueue = append(workQueue, branch{graph: node, tree: rootTree})

	// maju branch -> iterate
	for len(workQueue) > 0 {
		current := workQueue[0]
		workQueue = workQueue[1:]

		// If leaf -> append ke slice
		if Graph.IsLeaf(current.graph, graphMap) {
			result = append(result, Tree.CopyTree(current.tree))
			continue
		}

		// Iterate semua combination dari branch level yg sama
		for _, combi := range current.graph.Combinations {
			branchTree := Tree.CopyTree(current.tree)
			branchTree.FirstChild = Tree.NewTree(combi.FirstElement)
			branchTree.SecondChild = Tree.NewTree(combi.SecondElement)

			workQueue = append(workQueue, branch{graph: combi.FirstElement, tree: branchTree.FirstChild})
			workQueue = append(workQueue, branch{graph: combi.SecondElement, tree: branchTree.SecondChild})
		}
	}

	return result
}

func FindShortestPath(trees []*Tree.Tree) *Tree.Tree {
	if len(trees) == 0 {
		return nil
	}
	return trees[0]
}
