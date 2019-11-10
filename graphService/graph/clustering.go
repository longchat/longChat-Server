package graph

import (
	slog "log"

	_ "github.com/go-cq/cq"
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

func GetClusterByUserId(userid int64) (int, error) {
	cypher := "match (p) where u.UserId={0} return p.GroupId"
	rows, err := graphDb.Query(cypher, userid)
	if err != nil {
		return 0, err
	}
	rows.Next()
	var groupId int
	err = rows.Scan(&groupId)
	if err != nil {
		return 0, err
	}
	return groupId, err
}

func Clustering() {
	cypher := `match (p)-[l:Like]-(p2) return p.id,collect(p2.id) as nodes,collect(l.counts) as nodeScore order by  p.id asc`
	rows, err := graphDb.Query(cypher)
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
			slog.Fatalf("scan error: %v", err)
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
			if clusters[i].FatherCluster != nil {
				continue
			}
			if clusters[i].IsClosed {
				put(&clusters[i])
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
		setCluster(i, &data)
	}
}

var sum int

func setCluster(i int, c *Cluster) {
	if c.SubCluster == nil {
		graphDb.Query("match (p) where p.id={0} set p.GroupId={1}", c.id, i)
		//fmt.Printf("%d ", c.id)
		sum++
	} else {
		for _, data := range c.SubCluster {
			setCluster(i, data)
		}
	}
	return
}

func calculateParaWithCandicate(c *Cluster, candidate *Relation, candidate2 *Relation) {
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

//append one cluster to another big cluster
func join(father *Cluster, child *Cluster) {
	father.SubCluster = append(father.SubCluster, child)
	child.FatherCluster = father
}

//merge two clusters into one big cluster
func merge(ca *Cluster, cb *Cluster) {
	father := Cluster{
		id:         len(newClusters),
		SubCluster: []*Cluster{ca, cb},
	}
	newClusters = append(newClusters, father)
	ca.FatherCluster = &newClusters[father.id]
	cb.FatherCluster = &newClusters[father.id]
}

func put(c *Cluster) {
	father := Cluster{
		id:           len(newClusters),
		SubCluster:   []*Cluster{c},
		ParaRelation: c.ParaRelation,
	}
	newClusters = append(newClusters, father)
	c.FatherCluster = &newClusters[father.id]
}
