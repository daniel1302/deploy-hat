package main

import(

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"fmt"
)

// InputArgs is struct to keep input arguments
type InputArgs struct {
	OldAMI string
	NewAMI string
}

type InfrastructureAction interface {
    Commit(info *PipelineInfo)
	Rollback(info *PipelineInfo)
}

type ShortInstanceDesc struct {
	Id string
	InstanceType string
	KeyName string
	SubnetId string
	VpcId string
	SecurityGroupsIds []string
}

type PipelineInfo struct {
	instances []ShortInstanceDesc
}

type ListInstancesAction struct {
	Svc   *ec2.EC2
	Input *InputArgs
}

func (this ListInstancesAction) Commit(pipelineInfo *PipelineInfo) {
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
            &ec2.Filter{
                Name: aws.String("image-id"),
                Values: []*string{
                    aws.String(this.Input.OldAMI),
                },
            },
        },
	}

	result, err := this.Svc.DescribeInstances(input)
	if err != nil {
		panic(err)
	}

	for _, item := range result.Reservations {
		instance := item.Instances[0]

		sgIds := []string{}

		for _, sg := range instance.SecurityGroups {
			sgIds = append(sgIds, *sg.GroupId)
		}

		pipelineInfo.instances = append(pipelineInfo.instances, ShortInstanceDesc{
			Id: *instance.InstanceId,
			InstanceType: *instance.InstanceType,
			KeyName: *instance.KeyName,
			SubnetId: *instance.SubnetId,
			VpcId: *instance.VpcId,
			SecurityGroupsIds: sgIds,
		})
	}

	fmt.Println(pipelineInfo)
}

func (this ListInstancesAction) Rollback(info *PipelineInfo) {
	// Nothing to rollback
}

