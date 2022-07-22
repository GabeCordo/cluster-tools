package etl

func (d Database) run() {
	node := GetNodeInstance()
	go node.Start()

	for {

	}
}

func (d *Database) AddCluster(name string, cluster *Cluster) {
	if _, ok := d.mapper[name]; ok {

	}
}
