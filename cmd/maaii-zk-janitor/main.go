package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/sirupsen/logrus"
)

var zkUrl = "192.168.118.11:2181,192.168.118.12:2181,192.168.118.13:2181/maaii"
var debug = true
var dryRun = false

const (
	CLUSTER_MEMBERSHIP string = "cluster-membership"
	CLUSTER_DATA       string = "cluster-data"
)

func main() {
	var log = logrus.New()

	if debug {
		log.Level = logrus.DebugLevel
	}

	splitted := strings.Split(zkUrl, "/")

	zkRoot := ""
	if len(splitted) == 2 {
		zkRoot = fmt.Sprintf("/%s", splitted[1])
	}
	zkHosts := zk.FormatServers(strings.Split(splitted[0], ","))

	log.Infof("Connecting to %v using root=%s", zkHosts, zkRoot)
	zk.DefaultLogger = log
	c, _, err := zk.Connect(zkHosts, time.Second)
	if err != nil {
		log.Panic(err)
	}
	defer c.Close()

	members, _, err := c.Children(fmt.Sprintf("%s/%s", zkRoot, CLUSTER_MEMBERSHIP))
	if err != nil {
		log.Panic(err)
	}
	data_path := fmt.Sprintf("%s/%s", zkRoot, CLUSTER_DATA)
	data, _, err := c.Children(data_path)
	if err != nil {
		log.Panic(err)
	}

	// find nodes in data but not in members
	count := 0
	for _, d := range data {
		exist := false
		for _, m := range members {
			if d == m {
				exist = true
				break
			}
		}
		if !exist {
			log.Debugf("Found stale node: %s", d)
			count = count + 1

			// do not execute delete if running in dryRun mode
			if !dryRun {
				path := fmt.Sprintf("%s/%s", data_path, d)
				err = c.Delete(path, -1)
				if err != nil {
					log.Fatalf("Error deleting node %s, err=%v", path, err)
				}
			}
		}
	}

	if count > 0 {
		log.Infof("Removed %d stale nodes in %s", count, CLUSTER_DATA)

	} else {
		log.Info("No stale node found, exiting...")
	}
}
