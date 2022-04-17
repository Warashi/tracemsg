package main

import (
	"octracemsg"

	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(octracemsg.Analyzer) }
