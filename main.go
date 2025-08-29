package main

import (
	"fmt"
	"os"
	"encoding/json"
	"github.com/nostalgia296/asar-go/asar"
)

var (
    Version = "1.1"
    GoVersion string
    BuildAt string
    Author = "Nostalgia"
    jsonMap map[string]interface{}
)


func main() {
	args := os.Args

	// 检查参数数量是否足够
	if len(args) < 2 {
		fmt.Println("Usage: <command> <source> <destination>")
		fmt.Println("Commands:")
		fmt.Println("  p - Pack a directory into an ASAR file")
		fmt.Println("  e - Extract an ASAR file into a directory")
		fmt.Println("  v - Show version and buildtime")
		fmt.Println("  l - Show path and file in .asar")
		return
	}

	// 检查命令
	switch args[1] {
	case "p":
		if len(args) < 4 {
			fmt.Println("Usage: p <source directory> <output ASAR file>")
			return
		}
		// 调用 Pack 方法
		asar.Pack(args[2], args[3])
	case "e":
		if len(args) < 4 {
			fmt.Println("Usage: e <ASAR file> <output directory>")
			return
		}
		asar.Unpack(args[2], args[3])
	case "l":
	    if len(args) < 3 {
			fmt.Println("Usage: l <ASAR file> ")
			return
			}
        jsons, _ := asar.ReadJson(args[2])
	    err := json.Unmarshal([]byte(jsons), &jsonMap)
	    if err != nil {
		    fmt.Println("Error unmarshalling JSON")
	                 }
	    asar.Traverse(jsonMap["files"].(map[string]interface{}), "./")

    case "v":
        fmt.Printf("Version : %s \n", Version)
        fmt.Printf("BuildAt : %s \n", BuildAt)
        fmt.Printf("GoVersion : %s \n", GoVersion)
        fmt.Printf("Author : %s \n", Author)
	default:
		fmt.Println("Invalid command. Use 'p' to pack, 'e' to extract , 'v' to show version , 'l' to show list")
	}
}
