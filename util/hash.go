package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var txtPath string
var hashedPath string
var split string
var vertex int

func main() {
	txtPath = os.Args[1]
	hashedPath = os.Args[2]
	if os.Args[3] == "TAB" {
		split = "\t"
	} else {
		split = " "
	}
	vertex, _ = strconv.Atoi(os.Args[4])
	fIn, err := os.Open(txtPath)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(fIn)
	var mx, cnt int
	hash := make([]uint32, vertex*3)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if line[0] == '#' {
			continue
		}
		line = strings.TrimRight(line, "\r\n")
		nodes := strings.Split(line, split)
		u, _ := strconv.Atoi(nodes[0])
		v, err := strconv.Atoi(nodes[1])
		if u > mx {
			mx = u
		}
		if v > mx {
			mx = v
		}
		hash[u] = 1
		hash[v] = 1
		cnt++
		if cnt <= 50 {
			fmt.Println(u, v)
		}
		if cnt%1000000 == 0 {
			fmt.Println("Read", cnt, "Edges")
		}
	}
	for i := 1; i <= mx; i++ {
		hash[i] += hash[i-1]
	}
	for i := 0; i <= mx; i++ {
		hash[i]--
	}
	fmt.Println("Hash Complete")
	cnt = 0
	fIn.Seek(0, 0)
	reader = bufio.NewReader(fIn)
	fOut, err := os.Create(hashedPath)
	if err != nil {
		panic(err)
	}
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if line[0] == '#' {
			continue
		}
		line = strings.TrimRight(line, "\r\n")
		nodes := strings.Split(line, split)
		u, _ := strconv.Atoi(nodes[0])
		v, _ := strconv.Atoi(nodes[1])
		s := strconv.Itoa(int(hash[u])) + " " + strconv.Itoa(int(hash[v])) + "\n"
		fOut.WriteString(s)
		cnt++
		if cnt <= 50 {
			fmt.Println(u, v)
		}
		if cnt%1000000 == 0 {
			fmt.Println("Write", cnt, "Edges")
		}
	}
	fIn.Close()
	fOut.Close()
}
