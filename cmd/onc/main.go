package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

var version string

func main() {
	fmt.Printf("Starting onc, version %s\n", version)
	lambda.Start(calculatorHandler)
}
