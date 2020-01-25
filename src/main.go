package main

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ec2"
    "github.com/aws/aws-sdk-go/service/elbv2"
    "fmt"
    "os"
    "time"
)

func rollback(step int, pipelineInfo *PipelineInfo, actions *[]InfrastructureAction) {
    for step >= 0 {
        (*actions)[step].Rollback(pipelineInfo)
        fmt.Printf("[%T] Rolling changes back\n", (*actions)[step])
        step--
    }
}

func main() {
    if len(os.Args) != 3 {
        fmt.Printf("[ERROR] Invalid usage. usage: %s OLD_AMI NEW_AMI\n", os.Args[0])
        os.Exit(1)
    }

    sess, _ := session.NewSession(&aws.Config{
        Region: aws.String("us-east-1")},
    )
    
    svc := ec2.New(sess)
    elbv2 := elbv2.New(sess)
    pipelineInfo := &PipelineInfo{
        Version: time.Now().Format("20060102_150405"),
    }

    actions := []InfrastructureAction{
        InitializePipelineAction{os.Args[1], os.Args[2]},
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
            os.Exit(2)
        }

        fmt.Printf("[%T] Finished. No errors\n", action)
    }
}
