package workflow

// ==================== DAG 图结构 ====================
type dagGraph struct {
	nodes    map[string]*Node  // nodeID -> Node
	outEdges map[string][]Edge // nodeID -> 出边列表
	inEdges  map[string][]Edge // nodeID -> 入边列表
}

func buildGraph(wf *Workflow) *dagGraph {
	g := &dagGraph{
		nodes:    make(map[string]*Node),
		outEdges: make(map[string][]Edge),
		inEdges:  make(map[string][]Edge),
	}
	for i := range wf.Nodes {
		g.nodes[wf.Nodes[i].ID] = &wf.Nodes[i]
	}
	for _, edge := range wf.Edges {
		g.outEdges[edge.Source] = append(g.outEdges[edge.Source], edge)
		g.inEdges[edge.Target] = append(g.inEdges[edge.Target], edge)
	}
	return g
}
