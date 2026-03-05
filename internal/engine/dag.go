package engine

import (
	"fmt"

	"github.com/2comjie/mcpflow/internal/model"
)

type dagGraph struct {
	nodes    map[string]*model.Node
	outEdges map[string][]model.Edge // nodeID -> outgoing edges
	inEdges  map[string][]model.Edge // nodeID -> incoming edges
}

func buildGraph(nodes []model.Node, edges []model.Edge) (*dagGraph, error) {
	g := &dagGraph{
		nodes:    make(map[string]*model.Node, len(nodes)),
		outEdges: make(map[string][]model.Edge),
		inEdges:  make(map[string][]model.Edge),
	}

	for i := range nodes {
		n := &nodes[i]
		if _, exists := g.nodes[n.ID]; exists {
			return nil, fmt.Errorf("duplicate node id: %s", n.ID)
		}
		g.nodes[n.ID] = n
	}

	for _, e := range edges {
		if _, ok := g.nodes[e.Source]; !ok {
			return nil, fmt.Errorf("edge source %s not found", e.Source)
		}
		if _, ok := g.nodes[e.Target]; !ok {
			return nil, fmt.Errorf("edge target %s not found", e.Target)
		}
		g.outEdges[e.Source] = append(g.outEdges[e.Source], e)
		g.inEdges[e.Target] = append(g.inEdges[e.Target], e)
	}

	return g, nil
}

// findStartNode 找到 start 类型的节点
func (g *dagGraph) findStartNode() (*model.Node, error) {
	for _, n := range g.nodes {
		if n.Type == model.NodeStart {
			return n, nil
		}
	}
	return nil, fmt.Errorf("no start node found")
}

// getNextNodes 获取节点的下游节点，对条件节点按 branch 过滤
func (g *dagGraph) getNextNodes(nodeID string, branch string) []*model.Node {
	var next []*model.Node
	for _, e := range g.outEdges[nodeID] {
		if e.Condition != "" && e.Condition != branch {
			continue
		}
		if n, ok := g.nodes[e.Target]; ok {
			next = append(next, n)
		}
	}
	return next
}
