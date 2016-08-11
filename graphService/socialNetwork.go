package main

import (
	"flag"
	slog "log"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/graphService/graph"
)

func main() {
	pconfig := flag.String("config", "../config.cfg", "config file")
	psection := flag.String("section", "dev", "section of config file to apply")
	flag.Parse()
	config.InitConfig(pconfig, psection)
	url, err := config.GetConfigString(consts.Neo4JDbUrl)
	if err != nil {
		slog.Fatalln("get GetConfigString failed!:", consts.Neo4JDbUrl)
	}
	graph.NewDb(url)
	defer graph.FinDb()
	graph.Clustering()
}
