package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chankh/maaii-zk-janitor/pkg/version"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/sirupsen/logrus"
)

const (
	CLUSTER_MEMBERSHIP string = "cluster-membership"
	CLUSTER_DATA       string = "cluster-data"
)

func main() {
	fs := flag.NewFlagSet("", flag.ExitOnError)

	zkUrl := fs.String("zk", "localhost", "Zookeeper connection string")
	debug := fs.Bool("debug", false, "Enable debug logs")
	dryRun := fs.Bool("dry-run", false, "Dry run mode, does not actually delete nodes")
	showVer := fs.Bool("version", false, "Print the version number")

	flag.Usage = fs.Usage

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	if *showVer {
		fmt.Printf("Version\t\t: %s\n", version.Version)
		fmt.Printf("Build Date\t: %s\n", version.BuildDate)
		fmt.Printf("Build Number\t: %s\n", version.BuildNumber)
		fmt.Printf("Build Hash\t: %s\n", version.BuildHash)
		fmt.Printf("Platform\t: %s\n", version.BuildPlatform)
		fmt.Printf("Go Version\t: %s\n", version.GoVersion)
		os.Exit(0)
	}
	var log = logrus.New()

	if *debug {
		log.Level = logrus.DebugLevel
		log.Debug("Debug mode enabled")
	}

	splitted := strings.Split(*zkUrl, "/")

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
			if !*dryRun {
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
