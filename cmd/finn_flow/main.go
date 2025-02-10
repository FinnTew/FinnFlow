package main

import (
	"context"
	"fmt"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/cel"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/parser"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/smctx"
)

func main() {
	e := cel.NewEvaluator()
	r, err := e.Evaluate("0!=0", smctx.New(context.TODO()))
	if err != nil {
		panic(err)
	}
	fmt.Println(r)

	parser.NewParser(parser.NewDefaultRegistry())
}
