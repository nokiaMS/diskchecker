package main

import (
	"flag"
	"os"
	"bufio"
	"io"
	"fmt"
	"strings"
)

func LoadConfig(file string) []string {
	f,err := os.Open(file)
	if(err != nil) {
		panic(err)
	}
	defer f.Close()

	var config []string
	rd := bufio.NewReader(f)
	for {
		line,err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("File comes to end.")
				break
			}
			panic(err)
		}
		if line[0] != '#' && line[0] != '\n' {
			line = strings.Replace(line,"\n","",-1)
			config = append(config, line)
		}
	}
	return config
}

func IsIn( line string, list []string) bool {
	for _,v := range list {
		if strings.Contains(line,v) {
			return true
		}
	}
	return false
}

func main() {
	configFile := flag.String("config", "","Config file path.")
	srcFile := flag.String("src", "","Source file to be analyzed.")
	targetFile := flag.String("target", "","Target file to store the result.")
	flag.Parse()

	fmt.Printf("Config file: %s\n", *configFile)
	fmt.Printf("Source file: %s\n", *srcFile)
	fmt.Printf("Target file: %s\n", *targetFile)

	//open target file.
	target,err := os.OpenFile(*targetFile, os.O_CREATE|os.O_RDWR|os.O_APPEND ,0666)
	if(err != nil) {
		panic(err)
	}
	defer target.Close()
	wr := bufio.NewWriter(target)

	src,err := os.Open(*srcFile)
	if(err != nil) {
		panic(err)
	}
	defer src.Close()
	rd := bufio.NewReader(src)

	filterList := LoadConfig(*configFile)	//get filter list.
	for {
		line,err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
		}

		//write line to the new file.
		if IsIn(line, filterList) {
			fmt.Println(line)
			wr.WriteString(line)
		}
	}
}
