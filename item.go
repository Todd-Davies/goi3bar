package goi3bar

import (
	"fmt"
	"os"
	"time"
)

type Generator interface {
	Generate() (Output, error)
}

type Item struct {
	Generator

	Name    string
	out     chan<- Output
	refresh time.Duration
	kill    chan struct{}
}

func (i Item) Start() {
	if i.out == nil || i.kill == nil {
		panic("Item must be registered before starting")
	}

	go i.loop()
}

func (i *Item) loop() {
	t := time.NewTicker(i.refresh)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			output, err := i.Generate()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating output for %v: %v\n", i.Name, err)
				continue
			}

			// Try to asynchronously send the output, if it's time for another output pack, abandon it
			go func() {
				select {
				case <-time.After(i.refresh):
					return
				case i.out <- output:
					return
				}
			}()

		case <-i.kill:
			return
		}
	}
}

func NewItem(name string, interval time.Duration, g Generator) *Item {
	item := Item{
		Generator: g,

		Name:    name,
		out:     nil,
		refresh: interval,
		kill:    nil,
	}

	return &item
}
