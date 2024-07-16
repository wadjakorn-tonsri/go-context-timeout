package main

import (
	"context"
	"fmt"
)

type ContextKey string

const key ContextKey = "somekey"

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, key, make(chan int))
	valStr, ok := ctx.Value(key).(string)
	if !ok {
		fmt.Println("Value is not a string")
		return
	}
	fmt.Printf("Value: %s\n", valStr)
}
