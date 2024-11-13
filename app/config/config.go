package config

import "flag"

var (
	dir        string
	dbfilename string
	port       string

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
	flag.StringVar(&dir, "dir", "/tmp/redis-files", "directory where the RDB file is stored")
	flag.StringVar(&dbfilename, "dbfilename", "dump.rdb", "name of the RDB file")
	flag.StringVar(&port, "port", "6379", "port on which to listen")
	flag.Parse()
	Set("dir", dir)
	Set("dbfilename", dbfilename)
	Set("port", port)
}
