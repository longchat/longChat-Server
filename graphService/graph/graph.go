package graph

import (
	"fmt"
	slog "log"

	_ "github.com/go-cq/cq"
	"github.com/jmoiron/sqlx"
)

func StartAnalysis(neo4jURL string) {
	db, err := sqlx.Connect("neo4j-cypher", neo4jURL)
	if err != nil {
		slog.Fatalln("error connecting to neo4j:", err)
	}
	defer db.Close()

	cypher := `match (p)-[:Like]-(p2) return id(p),collect(id(p2)) as nodes order by id(p) asc`
	rows, err := db.Query(cypher)
	if err != nil {
		slog.Fatalln("error querying movie:", err)
	}
	defer rows.Close()
	baseNode := make([][]int, 0, 1500)
	baseId := -1
	for rows.Next() {
		var pId int
		var pIdArray []int
		err := rows.Scan(&pId, &pIdArray)
		if err != nil {
			fmt.Println(err)
		}
		if baseId == -1 {
			baseId = pId
		}
		baseNode = append(baseNode, pIdArray)
	}
	clusteringGraph := make([][]int, 0)
	clusteringNode := make(map[int]int)
	for i := range{
		
	}
}
