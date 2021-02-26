package main

import (
	"os"
	"strconv"
	"time"

	"github.com/AllFi/bft-demo/node"
)

func main() {
	baseDir := `C:\Users\aleks\Desktop\test`
	err := cleanUp(baseDir)

	persistentPeers := ""
	for i := 0; i < 4; i++ {
		id, address, err := node.InitNewNode(baseDir+`\node`+strconv.Itoa(i), i)
		if err != nil {
			panic(err)
		}
		persistentPeers += id + "@" + address + ","
	}

	for i := 0; i < 4; i++ {
		app := NewApplication(true, i)
		go node.RunNewNode(app, baseDir+`\node`+strconv.Itoa(i), i, persistentPeers)
	}

	for true {
		time.Sleep(1)
	}

	println(err)
}

func cleanUp(basePath string) (err error) {
	return os.RemoveAll(basePath)
}
