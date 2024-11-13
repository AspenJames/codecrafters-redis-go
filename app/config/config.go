package config

import "flag"

var (
	dbfilename string
	dir        string
	port       string
	replicaof  string

	config cfg = make(cfg)
)

type cfg map[string]string

func Get(key string) (string, bool) {
	v, ok := config[key]
	return v, ok
}

func Set(key, value string) {
	config[key] = value
}

func ParseCLIFlags() {
	flag.StringVar(&dbfilename, "dbfilename", "dump.rdb", "name of the RDB file")
	flag.StringVar(&dir, "dir", "/tmp/redis-files", "directory where the RDB file is stored")
	flag.StringVar(&port, "port", "", "port on which to listen")
	flag.StringVar(&replicaof, "replicaof", "", "<MASTER HOST> <MASTER PORT>")
	flag.Parse()
	Set("dir", dir)
	Set("dbfilename", dbfilename)
	Set("replicaof", replicaof)
	if port == "" {
		// Set default port; depends on replicaof status.
		if replicaof == "" {
			port = "6379"
		} else {
			port = "6380"
		}
	}
	Set("port", port)
}
