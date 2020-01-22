package main

import (
	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	// "fmt"
	// "os"
)



func main() {

	sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-1")},
    )

	svc := ec2.New(sess)
	inputArgs := &InputArgs{"ami-0c2db42c00f8ff366", "ami-0c2db42c00f8ff366"}
	list := ListInstancesAction{svc}
	pipelineInfo := &PipelineInfo{Input: inputArgs}

	_ = err
	list.Commit(pipelineInfo)
	// input := &ec2.DescribeInstancesInput{
	// 	Filters: []*ec2.Filter{
    //         &ec2.Filter{
    //             Name: aws.String("image-id"),
    //             Values: []*string{
    //                 aws.String("ami-0c2db42c00f8ff366"),
    //             },
    //         },
    //     },
	// }

	// result, err := svc.DescribeInstances(input)
	// if err != nil {
		
	// 	fmt.Println(err)
		
	// 	return
	// }

}