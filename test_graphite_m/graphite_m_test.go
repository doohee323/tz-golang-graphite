package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap(t *testing.T) {
	fmt.Println("===================================================")
	STYPE = "core"
	HCIDS = append(HCIDS, 1418, 1419, 2502, 2694, 2932, 2933, 2695)
	fmt.Println(HCIDS)
	result := mainExec()
	fmt.Println(result)
	fmt.Println(len(HCIDS))
	assert.Equal(t, result, len(HCIDS), "at least test should pass")

//	fmt.Println("===================================================")
//	STYPE = "graphite"
//	HCIDS = []int{}
//	HCIDS = append(HCIDS, 1418, 1419, 2502, 2694, 2932, 2933, 2695)
//	fmt.Println(HCIDS)
//	result = mainExec()
//	fmt.Println(result)
//	assert.Equal(t, result, len(HCIDS), "at least test should pass")
}
