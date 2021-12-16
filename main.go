package main

import (
	"log"

	jsoniter "github.com/json-iterator/go"
	"github.com/yoannduc/mergesync/internal/usecase/mergetype"
)

func main() {
	o := mergetype.MergedList(10)

	if j, err := jsoniter.MarshalToString(o); err == nil {
		log.Printf("o | %T | %v\n", j, j)
	} else {
		log.Printf("o | %T | %v\n", o, o)
	}

	log.Printf("len(o) | %T | %v\n", len(o), len(o))
}
