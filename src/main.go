package main

import (
	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"fmt"
	// "os"
	"time"
)

func rollback(step int, pipelineInfo *PipelineInfo, actions *[]InfrastructureAction) {
    for step >= 0 {
        (*actions)[step].Rollback(pipelineInfo)
        step--
    }
}

func main() {
	sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-1")},
    )
    _ = err
	svc := ec2.New(sess)
	elbv2 := elbv2.New(sess)
	pipelineInfo := &PipelineInfo{
		Version: time.Now().Format("20060102_150405"),
		NewInstancesIds: []*string{aws.String("i-0bfbcdd06a87a429b"), aws.String("i-0dc032337d8964e97")},
	}

    actions := []InfrastructureAction{
        InitializePipelineAction{"ami-0d279985b668e9b38", "ami-0aa2563dfc98ff16b"},
    	ListInstancesAction{svc},
	    FindLoadBalancerAction{elbv2},
    	RunInstancesAction{svc},
	    WaitUntilStatusOkAction{svc},
    	AuthorizeSecurityGroupsAction{svc},
	    CollectPublicIpsAction{svc},
    	TestInstancesAction{svc},
	    RegisterNewInstancesAction{elbv2},
    	DeregisterOldInstancesAction{elbv2},
    	WaitForDeregisterAction{elbv2},
	    TerminateOldInstancesAction{svc},
	
    }
    
    
    for idx, action := range actions {
        fmt.Printf("[%T] Executing.\n", action)
        err := action.Commit(pipelineInfo)

        if err != nil {
            fmt.Printf("[%T][ERROR] %s\n", action, err.Error())
            rollback(idx, pipelineInfo, &actions)
            break
        }

        fmt.Printf("[%T] Finished. No errors\n", action)
    }


}
