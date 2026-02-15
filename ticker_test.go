package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	after := time.After(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("done")
			return
		case <-ticker.C:
			fmt.Println("tick", time.Now().Format("2006-01-02 15:04:05"))
		case <-after:
			fmt.Println("timeout")
			return
		}
	}

}
