package graph

import (
	"fmt"
	slog "log"

	_ "github.com/go-cq/cq"
	"github.com/jmoiron/sqlx"
)

type Relation struct {
	id    int
	score int
}
type Cluster struct {
	id            int
	SubCluster    []*Cluster
	ParaRelation  []Relation
	FatherCluster *Cluster
	IsClosed      bool
}

var clusters []Cluster
var oldClusters []Cluster
var newClusters []Cluster

func StartAnalysis(neo4jURL string) {
	db, err := sqlx.Connect("neo4j-cypher", neo4jURL)
	if err != nil {
		slog.Fatalln("error connecting to neo4j:", err)
	}
	defer db.Close()

	cypher := `match (p)-[l:Like]-(p2) return p.id,collect(p2.id) as nodes,collect(l.counts) as nodeScore order by  p.id asc`
	rows, err := db.Query(cypher)
	if err != nil {
		slog.Fatalln("error querying movie:", err)
	}
	defer rows.Close()
	clusters = make([]Cluster, 0, 1500)
	baseId := -1
	cmap := make(map[int]bool)
	for rows.Next() {
		var cid int
		var pids []int
		var pscores []int
		err := rows.Scan(&cid, &pids, &pscores)
		if err != nil {
			fmt.Println(err)
		}
		_, isok := cmap[cid]
		if isok || cid > 1144 {
			slog.Fatalf("duplicate:%d  %v  \n", cid, pids)
		} else {
			cmap[cid] = true
		}
		if baseId == -1 {
			baseId = cid
		}
		var relations []Relation
		for i := range pids {
			if pids[i] > 1144 {
				fmt.Println(pids[i])
			}
			relations = append(relations, Relation{id: pids[i] - baseId, score: pscores[i]})
		}
		oneCluster := Cluster{id: len(clusters), ParaRelation: relations}

		clusters = append(clusters, oneCluster)
	}
	oldClusters = make([]Cluster, len(clusters))
	newClusters = make([]Cluster, 0, 700)
	idx := 0
	for idx < 6 {
		idx++
		for i := range clusters {
			if clusters[i].FatherCluster != nil || clusters[i].IsClosed {
				continue
			}
			if clusters[i].SubCluster == nil {
				candidate := Relation{score: -1, id: -1}
				for j := range clusters[i].ParaRelation {
					p := clusters[i].ParaRelation[j]
					if p.score > candidate.score && clusters[p.id].FatherCluster == nil {
						candidate = p
					}
				}
				if candidate.id == -1 {
					for j := range clusters[i].ParaRelation {
						p := clusters[i].ParaRelation[j]
						if p.score > candidate.score {
							candidate = p
						}
					}
					join(clusters[candidate.id].FatherCluster, &clusters[i])
				} else {
					merge(&clusters[i], &clusters[candidate.id])
				}
			} else {
				candidate := Relation{score: -1, id: -1}
				candidate2 := Relation{score: -1, id: -1}
				calculateParaWithCandicate(&clusters[i], &candidate, &candidate2)
				if candidate.id != -1 {
					calculatePara(&clusters[candidate.id])
					merge(&clusters[i], &clusters[candidate.id])
				} else if candidate2.id != -1 {
					join(clusters[candidate2.id].FatherCluster, &clusters[i])
				} else {
					clusters[i].IsClosed = true
					put(&clusters[i])
				}
			}
		}
		oldClusters = make([]Cluster, len(clusters))
		copy(oldClusters, clusters)
		clusters = make([]Cluster, len(newClusters))
		copy(clusters, newClusters)
		newClusters = make([]Cluster, 0, len(newClusters)/2)
	}

	for i, data := range clusters {
		fmt.Print(i, ":\n")
		printCluster(&data)
		fmt.Print("\n\n")
	}
}

func printCluster(c *Cluster) {
	if c.SubCluster == nil {
		fmt.Print(" ", c.id)
	} else {
		for _, data := range c.SubCluster {
			printCluster(data)
		}
	}
	return
}

func calculateParaWithCandicate(c *Cluster, candidate *Relation, candidate2 *Relation) {
	paraRelation := make(map[int]Relation)
	for j := range c.SubCluster {
		sub := c.SubCluster[j]
		if len(sub.ParaRelation) == 0 {
			slog.Println("null para", sub.id)
		}

		for k := range sub.ParaRelation {
			fId := oldClusters[sub.ParaRelation[k].id].FatherCluster.id
			if fId != c.id {
				r, isok := paraRelation[fId]
				if !isok {
					r = Relation{id: fId, score: sub.ParaRelation[k].score}
				} else {
					r.score = sub.ParaRelation[k].score + r.score
				}
				paraRelation[fId] = r
				if r.score > candidate.score {
					if clusters[fId].FatherCluster == nil {
						*candidate = r
					}
					*candidate2 = r
				}
			}
		}
	}
	for _, v := range paraRelation {
		c.ParaRelation = append(c.ParaRelation, v)
	}
}

func calculatePara(c *Cluster) {
	paraRelation := make(map[int]Relation)
	for j := range c.SubCluster {
		sub := c.SubCluster[j]
		for k := range sub.ParaRelation {
			fId := oldClusters[sub.ParaRelation[k].id].FatherCluster.id
			if fId != c.id {
				r, isok := paraRelation[fId]
				if !isok {
					r = Relation{id: fId, score: sub.ParaRelation[k].score}
				} else {
					r.score = sub.ParaRelation[k].score + r.score
				}
				paraRelation[fId] = r
			}
		}
	}
	for _, v := range paraRelation {
		c.ParaRelation = append(c.ParaRelation, v)
	}
}

func join(father *Cluster, child *Cluster) {
	father.SubCluster = append(father.SubCluster, child)
	child.FatherCluster = father
}

func merge(ca *Cluster, cb *Cluster) {
	father := Cluster{
		id:         len(newClusters),
		SubCluster: []*Cluster{ca, cb},
	}
	newClusters = append(newClusters, father)
	ca.FatherCluster = &father
	cb.FatherCluster = &father
}

func put(c *Cluster) {
	c.id = len(newClusters)
	newClusters = append(newClusters, *c)
}
