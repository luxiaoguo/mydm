package main

import (
	"fmt"
	"time"
)

type ceshi struct {
	dm     []string
	thisdm string
}

func (c *ceshi) deal() {
	for _, v := range c.dm {
		c.thisdm = v
		go c.print(c.thisdm)
	}
}

func (c *ceshi) print(thisdm string) {
	fmt.Println(thisdm)
}

func main() {
	var slc = []string{"aaa", "bbb", "ccc"}
	cs := &ceshi{dm: slc}
	css := &ceshi{dm: slc}
	var cslc []*ceshi
	cslc = append(cslc, cs)
	cslc = append(cslc, css)
	for _, v := range cslc {
		v.deal()
	}
	time.Sleep(5 * time.Second)
}
