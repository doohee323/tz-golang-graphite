package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap(t *testing.T) {
	fmt.Println("===================================================")
	result := mainExec()
	assert.Equal(t, result, 1, "at least test should pass")
	
	fmt.Println("===================================================")
	result = mainExec()
	assert.Equal(t, result, 1, "at least test should pass")
}
