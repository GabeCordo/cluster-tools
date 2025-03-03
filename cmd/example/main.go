package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sentmint/pops"
	"github.com/Sentmint/yule"
)

type Injectible struct {
	Message string
}

func extract(inj Injectible, out chan string) {

	fmt.Println(inj.Message)

	// assumption: data is pulled from database and pushed to transform in another action
	for i := 0; i < 1000000; i++ {
		out <- "foo"
	}

	time.Sleep(20 * time.Second)

	for i := 0; i < 500000; i++ {
		out <- "bar"
	}

	close(out)
}

func transform(inj Injectible, in string) (out string, err error) {

	// "foo" and "bar" are the only valid types
	if (in != "foo") && (in != "bar") {
		return "", errors.New("'in' must be of value ('foo' or 'bar')")
	}

	// assumption: processing some unit of data takes 4ms
	time.Sleep(4 * time.Millisecond)

	return in, nil
}

func load(inj Injectible, in string) {

	// assumption: uploading to database takes 3ms
	time.Sleep(3 * time.Millisecond)
}

var Pipeline yule.RunnablePipeline = yule.Build(yule.F{Value: extract}, yule.F{Value: transform}, yule.F{Value: load})

func main() {

	start := time.Now()

	inj := Injectible{Message: "hello there"}

	go func() {
		for {
			Pipeline.Stats.Print()

			time.Sleep(100 * time.Millisecond)
		}
	}()

	cfg := &pops.Config{}
	cfg.Core.Host = "https://localhost:8137/"
	cfg.Core.Attempts = 10

	pops.Run(cfg, &Pipeline, inj)

	Pipeline.Stats.Print()

	end := time.Now()

	fmt.Printf("pipelined took %d seconds\n", int(end.Sub(start).Seconds()))
}
