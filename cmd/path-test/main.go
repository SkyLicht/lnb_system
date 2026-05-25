package main

import (
	"flag"
	"fmt"
	"os"

	"lnb/utils"
)

func main() {
	path := flag.String("path", "", "path to test for existence")
	network := flag.Bool("network", false, "test path as a network UNC path")
	connect := flag.Bool("connect", false, "try to connect the network share before testing")
	user := flag.String("user", "", "network share user")
	pass := flag.String("pass", "", "network share password")
	flag.Parse()

	if *network {
		runNetworkPathTest(*path, *connect, *user, *pass)
		return
	}

	status, err := utils.TestPathExistence(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "path test failed: %v\n", err)
		os.Exit(2)
	}

	if !status.Found {
		fmt.Fprintf(os.Stderr, "path does not exist: %s\n", status.Path)
		os.Exit(1)
	}

	kind := "file"
	if status.IsDir {
		kind = "directory"
	}

	fmt.Printf("path exists: %s\n", status.Path)
	fmt.Printf("type: %s\n", kind)
	fmt.Printf("size: %d\n", status.Size)
}

func runNetworkPathTest(path string, connect bool, user string, pass string) {
	status, err := utils.TestNetworkPathExistence(utils.NetworkPathOptions{
		Path:           path,
		Connect:        connect,
		UseCredentials: user != "" || pass != "",
		User:           user,
		Pass:           pass,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "network path test failed: %v\n", err)
		os.Exit(2)
	}

	if !status.Found {
		fmt.Fprintf(os.Stderr, "network path does not exist: %s\n", status.Path)
		fmt.Fprintf(os.Stderr, "share: %s\n", status.Share)
		os.Exit(1)
	}

	kind := "file"
	if status.IsDir {
		kind = "directory"
	}

	fmt.Printf("network path exists: %s\n", status.Path)
	fmt.Printf("share: %s\n", status.Share)
	fmt.Printf("type: %s\n", kind)
	fmt.Printf("size: %d\n", status.Size)
	fmt.Printf("connected: %t\n", status.Connected)
}
