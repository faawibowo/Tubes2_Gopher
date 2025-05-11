export type Tree = {
  first: TreeNodeElement;
};
export type TreeNodeElement = {
  name: string;
  children?: TreeNodeRecipe[];
};

export type TreeNodeRecipe = {
  firstElement: TreeNodeElement;
  secondElement: TreeNodeElement;
};

export type TreantNode = {
  text: { name: string };
  HTMLclass?: string;
  children?: TreantNode[];
};

export type TreantChart = {
  chart: {
    container: string;
    connectors: { type: string };
    node: { HTMLclass: string };
  };
  nodeStructure: TreantNode;
};

export function convertToTreant(tree: Tree): TreantChart {
  function buildNode(node: TreeNodeElement): TreantNode {
    // Leaf node
    if (!node.children || node.children.length === 0) {
      return {
        text: { name: node.name },
      };
    }

    const recipeNodes = node.children.map((recipe: TreeNodeRecipe) => {
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
