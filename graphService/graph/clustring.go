package graph

import (
	"fmt"
	slog "log"

	_ "github.com/go-cq/cq"
	"github.com/jmoiron/sqlx"
)

type Node struct {
	id    int
	score int
}

type Nodes []int

func StartAnalysis(neo4jURL string) {
	db, err := sqlx.Connect("neo4j-cypher", neo4jURL)
	if err != nil {
		slog.Fatalln("error connecting to neo4j:", err)
	}
	defer db.Close()

	cypher := `match (p)-[l:Like]-(p2) return id(p),collect(id(p2)) as nodes,collect(l.counts) as nodeScore order by id(p) asc`
	rows, err := db.Query(cypher)
	if err != nil {
		slog.Fatalln("error querying movie:", err)
	}
	defer rows.Close()

	baseNode := make([][]Node, 0, 1500)
	baseId := -1

	for rows.Next() {
		var pId int
		var pIdArray []int
		var pIdScore []int
		err := rows.Scan(&pId, &pIdArray, &pIdScore)
		if err != nil {
			fmt.Println(err)
		}

		if baseId == -1 {
			baseId = pId
		}
		nodes := make([]Node, len(pIdArray))
		for i := range pIdArray {
			pIdArray[i] = pIdArray[i] - baseId
			nodes[i] = Node{id: pIdArray[i], score: pIdScore[i]}
		}
		baseNode = append(baseNode, nodes)
	}

	clusteringGraph := make([]Nodes, 0)
	clusterNode := make([]int, len(baseNode))
	for i := range clusterNode {
		clusterNode[i] = -1
	}
	for i := range baseNode {
		if clusterNode[i] != -1 {
			break
		}
		maxNode := Node{id: -1}
		for j := range baseNode[i] {
			if maxNode.score < baseNode[i][j].score && clusterNode[baseNode[i][j].id] == -1 {
				maxNode = baseNode[i][j]
			}
		}
		if maxNode.id == -1 {
			for j := range baseNode[i] {
				if maxNode.score < baseNode[i][j].score {
					maxNode = baseNode[i][j]
				}
			}
			clusterId := clusterNode[maxNode.id]
			clusteringGraph[clusterId] = append(clusteringGraph[clusterId], i)
			clusterNode[i] = clusterId
			break
		}
		clusterNode[i] = len(clusteringGraph) - 1
		clusterNode[maxNode.id] = len(clusteringGraph) - 1
		clusteringGraph = append(clusteringGraph, []int{i, maxNode.id})
	}
	newclusterNode := make([]int, len(baseNode))
	for i := range clusterNode {
		newclusterNode[i] = -1
	}
	for i := range clusteringGraph {
		if newclusterNode[i] != -1 {
			break
		}
	}
}
