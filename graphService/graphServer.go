package main

import (
	"github.com/longchat/longChat-Server/graphService/graph"
)

func main() {
	graph.StartAnalysis("http://localhost:7474")
}
