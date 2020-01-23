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
	SecurityGroupsIds []*string
	Tags map[string]string
}

type PipelineInfo struct {
	Version string
	Input *InputArgs
	ClientIP string
	OldInstances []ShortInstanceDesc
	NewInstances []*string
	ModifiedSecurityGroups []*string
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

type WaitUntilStatusOkAction struct {
	Svc   *ec2.EC2
}

type AuthorizeSecurityGroups struct {
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
	clientIP, err := getClientIP()

	if err != nil {
		panic("TODO 4")
	}

	pipelineInfo.ClientIP = clientIP
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
			&ec2.Filter{
                Name: aws.String("instance-state-name"),
                Values: []*string{
                    aws.String("running"),
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

		sgIds := []*string{}

		for _, sg := range instance.SecurityGroups {
			sgIds = append(sgIds, sg.GroupId)
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

	if len(pipelineInfo.OldInstances) < 1 {
		panic("TODO 2")
	}
}

func (this ListInstancesAction) Rollback(pipelineInfo *PipelineInfo) {
	// Nothing to rollback
}

func (this RunInstancesAction) Commit(pipelineInfo *PipelineInfo) {
	for _, item := range pipelineInfo.OldInstances {
		oldTags := item.Tags
		newTags := []*ec2.Tag{}
		oldTags["Version"] = pipelineInfo.Version
		
		for tagKey, tagVal := range oldTags {
			newTags = append(newTags, &ec2.Tag{Key: aws.String(tagKey), Value: aws.String(tagVal)})
		}

		input := &ec2.RunInstancesInput{


			ImageId:      aws.String(pipelineInfo.Input.NewAMI),
			InstanceType: aws.String(item.InstanceType),
			KeyName:      aws.String(item.KeyName),
			MaxCount:     aws.Int64(1),
			MinCount:     aws.Int64(1),
			SecurityGroupIds: item.SecurityGroupsIds,
			SubnetId: aws.String(item.SubnetId),
			TagSpecifications: []*ec2.TagSpecification{
				{
					ResourceType: aws.String("instance"),
					Tags: newTags,
				},
			},
		}
		
		result, err := this.Svc.RunInstances(input)
		if err != nil {
			panic(err)
		}
		
		pipelineInfo.NewInstances = append(pipelineInfo.NewInstances, result.Instances[0].InstanceId)
	}	
}

func (this RunInstancesAction) Rollback(pipelineInfo *PipelineInfo) {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: pipelineInfo.NewInstances,
	}

	_, err := this.Svc.TerminateInstances(input)
	if err != nil {
		panic(err)
	}
	// fmt.Println(result)
}


func (this WaitUntilStatusOkAction) Commit(pipelineInfo *PipelineInfo) {

	input := &ec2.DescribeInstancesInput{
		InstanceIds: pipelineInfo.NewInstances,
	}

	err := this.Svc.WaitUntilInstanceRunning(input)

	if err != nil {
		panic(err)
	}

	
}

func (this WaitUntilStatusOkAction) Rollback(pipelineInfo *PipelineInfo) {
}

func (this AuthorizeSecurityGroups) Commit(pipelineInfo *PipelineInfo) {

	for _, instance := range pipelineInfo.OldInstances {
		if len(instance.SecurityGroupsIds) < 1 {
			panic("Instance must have Security Group")
		}

		isIpAuthorizedStatus, err := isIpAuthorized(this.Svc, *instance.SecurityGroupsIds[0], 80, pipelineInfo.ClientIP)
		if err != nil {
			panic("TODo 5")
		}

		if !isIpAuthorizedStatus {
			authorizeIp(this.Svc, *instance.SecurityGroupsIds[0], 80, pipelineInfo.ClientIP)
			pipelineInfo.ModifiedSecurityGroups = append(pipelineInfo.ModifiedSecurityGroups, instance.SecurityGroupsIds[0])
		}
	}
}

func (this AuthorizeSecurityGroups) Rollback(pipelineInfo *PipelineInfo) {
	for _, sgId := range pipelineInfo.ModifiedSecurityGroups {
		revokeIp(this.Svc, *sgId, 80, pipelineInfo.ClientIP)
	}
}