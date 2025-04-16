package testfunc

import "github.com/aws/aws-lambda-go/lambda"

func main() {
	lambda.Start(HandleRequest)
}
