package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap(t *testing.T) {
	all = append(all, 1418, 1419, 2502, 2694, 2932, 2933, 2695)
	fmt.Println(all)
	result := _main()
	fmt.Println(result)
	assert.Equal(t, result, len(all), "at least test should pass")
}
