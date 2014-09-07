package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/VonC/gogitolite/reader"
)

func main() {

	flag.Parse()
	filenames := flag.Args()
	if len(filenames) == 0 {
		fmt.Println("At least one gitolite.conf file expected")
		os.Exit(1)
	}

	for _, filename := range filenames {
		fmt.Printf("Read file '%v'\n", filename)
		f, err := os.Open(filename)
		if err != nil {
			fmt.Printf("ERR %v\n", err.Error())
			os.Exit(1)
		}
		defer f.Close()
		fr := bufio.NewReader(f)
		gtl, err := reader.Read(fr)
		if err != nil {
			fmt.Printf("ERR %v\n", err.Error())
			os.Exit(1)
		}
		// fmt.Println(gtl.String())
		for _, config := range gtl.Configs() {
			for _, rule := range config.Rules() {
				if rule.Access() != "-" {
					fmt.Println(config.GetRepos(), "=>", rule.GetUsers())
				}
			}
		}
	}
}
