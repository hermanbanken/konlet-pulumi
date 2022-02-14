package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		switch ctx.Stack() {
		// you can differentiate and run different stacks here!
		default:
			return stackExample(ctx)
		}
	})
}