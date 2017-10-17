package main

import (
	"github.com/jhunt/vcaptive"
)

type DummySystem struct {}

func (s DummySystem) Configure(services vcaptive.Services) (bool, error) {
	return true, nil
}

func (s DummySystem) Setup() error {
	return nil
}

func (s DummySystem) Summarize() Data {
	return Data{
		System: "Dummy",
		Summary: "*no summary*\n",
		Verification: "decafbad",
	}
}

func init() {
	if Systems == nil {
		Systems = make(map[string]System)
	}
	Systems["dummy"] = DummySystem{}
}
