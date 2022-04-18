package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/Warashi/tracemsg/octracemsg"
)

func main() { unitchecker.Main(octracemsg.Analyzer) }
