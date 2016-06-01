package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap(t *testing.T) {
	fmt.Println("===================================================")
	_type = "core"
	_hcids = append(_hcids, 1418, 1419, 2502, 2694, 2932, 2933, 2695)
	fmt.Println(_hcids)
	result := _main()
	fmt.Println(result)
	fmt.Println(len(_hcids))
	assert.Equal(t, result, len(_hcids), "at least test should pass")

//	fmt.Println("===================================================")
//	_type = "graphite"
//	_hcids = append(_hcids, 1418, 1419, 2502, 2694, 2932, 2933, 2695)
//	fmt.Println(_hcids)
//	result := _main()
//	fmt.Println(result)
//	assert.Equal(t, result, len(_hcids), "at least test should pass")
}
