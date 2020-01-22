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
	Tags map[string]string
}

type PipelineInfo struct {
	Version string
	Input *InputArgs
	OldInstances []ShortInstanceDesc
	NewInstances []string
}

type InitializePipelineAction struct {
	OldAMI string
	NewAMI string
}

type ListInstancesAction struct {
	Svc   *ec2.EC2
}

type RunInstancesAction struct {
	Svc   *ec2.EC2
}

func (this InitializePipelineAction) Commit(pipelineInfo *PipelineInfo) {
	OldAMI := this.OldAMI
	NewAMI := this.NewAMI
	fmt.Println( )

	if OldAMI[:4] != "ami-" || NewAMI[:4] != "ami-" {
		panic("TODO")	
	}

	if OldAMI == NewAMI {
		panic("TODO 1")
	}
	
	pipelineInfo.Input = &InputArgs{OldAMI, NewAMI}
}

func (this ListInstancesAction) Commit(pipelineInfo *PipelineInfo) {
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
            &ec2.Filter{
                Name: aws.String("image-id"),
                Values: []*string{
                    aws.String(pipelineInfo.Input.OldAMI),
                },
            },
        },
	}

	result, err := this.Svc.DescribeInstances(input)
	if err != nil {
		panic(err)
	}


	for _, item := range result.Reservations {
		fmt.Println(item) 
		instance := item.Instances[0]

		sgIds := []string{}

		for _, sg := range instance.SecurityGroups {
			sgIds = append(sgIds, *sg.GroupId)
		}

		tags := map[string]string{}

		for _, tag := range instance.Tags {
			tags[*tag.Key] = *tag.Value
		}

		pipelineInfo.OldInstances = append(pipelineInfo.OldInstances, ShortInstanceDesc{
			Id: *instance.InstanceId,
			InstanceType: *instance.InstanceType,
			KeyName: *instance.KeyName,
			SubnetId: *instance.SubnetId,
			VpcId: *instance.VpcId,
			SecurityGroupsIds: sgIds,
			Tags: tags,
		})
	}
}

func (this ListInstancesAction) Rollback(pipelineInfo *PipelineInfo) {
	// Nothing to rollback
}

func (this RunInstancesAction) Commit(pipelineInfo *PipelineInfo) {

	// fmt.Println(pipelineInfo)

	// for _, item := range pipelineInfo.OldInstances {


	// 	input := &ec2.RunInstancesInput{
	// 		ImageId:      aws.String(pipelineInfo.Input.NewAMI),
	// 		InstanceType: aws.String(item.InstanceType),
	// 		KeyName:      aws.String(item.KeyName),
	// 		MaxCount:     aws.Int64(1),
	// 		MinCount:     aws.Int64(1),
	// 		SecurityGroupIds: []*string{
	// 			aws.String(item.SecurityGroupsIds[0]),
	// 		},
	// 		SubnetId: aws.String(item.SubnetId),
	// 		TagSpecifications: []*ec2.TagSpecification{
	// 			{
	// 				ResourceType: aws.String("instance"),
	// 				Tags: []*ec2.Tag{
	// 					{
	// 						Key:   aws.String("Name"),
	// 						Value: aws.String("test-daniel"),
	// 					},
						
	// 				},
	// 			},
	// 		},
	// 	}
	// 	result, err := this.Svc.RunInstances(input)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println(result)
	// 	break
	// }
	
}

func (this RunInstancesAction) Rollback(pipelineInfo *PipelineInfo) {
	
}