package main

import (
	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"fmt"
	// "os"
	"time"
)



func main() {
	sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-1")},
    )

	svc := ec2.New(sess)
	pipelineInfo := &PipelineInfo{
		Version: time.Now().Format("20060102_150405"),
		// NewInstances: []*string{aws.String("i-04b257fdda45fffba")},
	}
	



	_ = err
	InitializePipelineAction{"ami-0c2db42c00f8ff366", "ami-08724dcf4c591e2f4"}.Commit(pipelineInfo)
	ListInstancesAction{svc}.Commit(pipelineInfo)
	RunInstancesAction{svc}.Commit(pipelineInfo)
	WaitUntilStatusOkAction{svc}.Commit(pipelineInfo)
	AuthorizeSecurityGroups{svc}.Commit(pipelineInfo)
	// RunInstancesAction{svc}.Rollback(pipelineInfo)


	fmt.Println(pipelineInfo)
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