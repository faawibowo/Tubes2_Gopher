export function convertToTreant(tree: any): any {
  function buildNode(node: any): any {
    // Leaf node
    if (!node.children || node.children.length === 0) {
      return {
        text: { name: node.name },
      };
    }

    const recipeNodes = node.children.map((recipe: any, i: number) => {
      return {
        text: { name: "Recipe" },
        HTMLclass: "recipeNode",
        children: [
          buildNode(recipe.firstElement),
          buildNode(recipe.secondElement),
        ],
      };
    });

    return {
      text: { name: node.name },
      children: recipeNodes,
    };
  }

  return {
    chart: {
      container: "#tree",
      connectors: { type: "step" },
      node: { HTMLclass: "nodeExample1" },
    },
    nodeStructure: buildNode(tree.first),
  };
}
