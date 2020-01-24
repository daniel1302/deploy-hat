package main

import(

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"

	"fmt"
	"errors"
)

// InputArgs is struct to keep input arguments
type InputArgs struct {
	OldAMI string
	NewAMI string
}

type InfrastructureAction interface {
    Commit(info *PipelineInfo) error
	Rollback(info *PipelineInfo) error
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
	OldInstancesIds []*string
	OldInstances []ShortInstanceDesc
	NewInstancesIds []*string
	NewInstancesIps []string
	ModifiedSecurityGroups []*string
	TargetGroupsArns []*string
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

type AuthorizeSecurityGroupsAction struct {
	Svc   *ec2.EC2
}

type TestInstancesAction struct {
	Svc   *ec2.EC2
}

type CollectPublicIpsAction struct {
	Svc   *ec2.EC2
}

type FindLoadBalancerAction struct {
	Svc   *elbv2.ELBV2
}

type RegisterNewInstancesAction struct {
	Svc   *elbv2.ELBV2
}

type WaitForDeregisterAction struct {
	Svc   *elbv2.ELBV2
}

type DeregisterOldInstancesAction struct {
	Svc   *elbv2.ELBV2
}

type TerminateOldInstancesAction struct {
	Svc   *ec2.EC2
}



func (this InitializePipelineAction) Commit(pipelineInfo *PipelineInfo) error {
	OldAMI := this.OldAMI
	NewAMI := this.NewAMI

	if len(OldAMI) < 4|| len(NewAMI) < 4 {
		return errors.New("Invalid AMI ID.")
	}

	if OldAMI[:4] != "ami-" || NewAMI[:4] != "ami-" {
		return errors.New("Invalid AMI ID.")
	}

	if OldAMI == NewAMI {
		return errors.New("Both new and old AMI ID are the same.")
	}

	pipelineInfo.Input = &InputArgs{OldAMI, NewAMI}
	clientIP, err := getClientIP()

	if err != nil {
		return err
	}

	pipelineInfo.ClientIP = clientIP
	return nil
}

func (this InitializePipelineAction) Rollback(pipelineInfo *PipelineInfo) error {
	return nil
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

		pipelineInfo.OldInstancesIds = append(pipelineInfo.OldInstancesIds, instance.InstanceId)
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
			NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
				{
					AssociatePublicIpAddress: aws.Bool(true),
					DeviceIndex:              aws.Int64(0),
					SubnetId:                 aws.String(item.SubnetId),
					Groups:					  item.SecurityGroupsIds,
				},
			},
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
		
		pipelineInfo.NewInstancesIds = append(pipelineInfo.NewInstancesIds, result.Instances[0].InstanceId)
	}
}

func (this RunInstancesAction) Rollback(pipelineInfo *PipelineInfo) {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: pipelineInfo.NewInstancesIds,
	}

	_, err := this.Svc.TerminateInstances(input)
	if err != nil {
		panic(err)
	}
	// fmt.Println(result)
}


func (this WaitUntilStatusOkAction) Commit(pipelineInfo *PipelineInfo) {

	input := &ec2.DescribeInstancesInput{
		InstanceIds: pipelineInfo.NewInstancesIds,
	}
	fmt.Println(input)
	err := this.Svc.WaitUntilInstanceRunning(input)

	if err != nil {
		panic(err)
	}	
}

func (this WaitUntilStatusOkAction) Rollback(pipelineInfo *PipelineInfo) {
}

func (this AuthorizeSecurityGroupsAction) Commit(pipelineInfo *PipelineInfo) {

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

func (this AuthorizeSecurityGroupsAction) Rollback(pipelineInfo *PipelineInfo) {
	for _, sgId := range pipelineInfo.ModifiedSecurityGroups {
		revokeIp(this.Svc, *sgId, 80, pipelineInfo.ClientIP)
	}
}


func (this CollectPublicIpsAction) Commit(pipelineInfo *PipelineInfo) {

	input := &ec2.DescribeInstancesInput{
		InstanceIds: pipelineInfo.NewInstancesIds,
		Filters: []*ec2.Filter{
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

		if *instance.PublicIpAddress == "" {
			panic("TODO 6")
		}

		pipelineInfo.NewInstancesIps = append(pipelineInfo.NewInstancesIps, *instance.PublicIpAddress)
	}
}

func (this CollectPublicIpsAction) Rollback(pipelineInfo *PipelineInfo) {
}



func (this TestInstancesAction) Commit(pipelineInfo *PipelineInfo) {
	for _, ip := range pipelineInfo.NewInstancesIps {
		isValid, err := isValidRequest("http://"+ip, 5)

		if err != nil {
			fmt.Println(err)
			panic("TODO 7")
		}

		if !isValid {
			panic("TODO 8")
		}
	}
}

func (this TestInstancesAction) Rollback(pipelineInfo *PipelineInfo) {
}




func (this FindLoadBalancerAction) Commit(pipelineInfo *PipelineInfo) {
	input := &elbv2.DescribeTargetGroupsInput{
		PageSize: aws.Int64(400),
	}

	res, err := this.Svc.DescribeTargetGroups(input)

	if err != nil {
		panic(err)
	}

	for _, tg := range res.TargetGroups {
		status, err := findInstancesInTargetGroup(this.Svc, *tg.TargetGroupArn, pipelineInfo.OldInstancesIds)

		if err != nil {
			panic(err)
		}
		if status == true {
			pipelineInfo.TargetGroupsArns = append(pipelineInfo.TargetGroupsArns, tg.TargetGroupArn)
		}
	}
}

func (this FindLoadBalancerAction) Rollback(pipelineInfo *PipelineInfo) {
	
}


func (this RegisterNewInstancesAction) Commit(pipelineInfo *PipelineInfo) {
	targets := []*elbv2.TargetDescription{}

	for _, instanceId := range pipelineInfo.NewInstancesIds {
		targets = append(targets, &elbv2.TargetDescription{Id: instanceId})
	}

	for _, tgArn := range pipelineInfo.TargetGroupsArns {
		input := &elbv2.RegisterTargetsInput{
			TargetGroupArn: tgArn,
			Targets: targets,
		}
		
		_, err := this.Svc.RegisterTargets(input)

		if err != nil {
			panic(err)
		}
	}
}

func (this RegisterNewInstancesAction) Rollback(pipelineInfo *PipelineInfo) {
	targets := []*elbv2.TargetDescription{}

	for _, instanceId := range pipelineInfo.NewInstancesIds {
		targets = append(targets, &elbv2.TargetDescription{Id: instanceId})
	}

	for _, tgArn := range pipelineInfo.TargetGroupsArns {
		input := &elbv2.DeregisterTargetsInput{
			TargetGroupArn: tgArn,
			Targets: targets,
		}
		
		_, err := this.Svc.DeregisterTargets(input)

		if err != nil {
			panic(err)
		}

		fmt.Println(input)
	}
}

func (this DeregisterOldInstancesAction) Commit(pipelineInfo *PipelineInfo) {
	targets := []*elbv2.TargetDescription{}

	for _, instanceId := range pipelineInfo.OldInstancesIds {
		targets = append(targets, &elbv2.TargetDescription{Id: instanceId})
	}

	for _, tgArn := range pipelineInfo.TargetGroupsArns {
		input := &elbv2.DeregisterTargetsInput{
			TargetGroupArn: tgArn,
			Targets: targets,
		}
		
		_, err := this.Svc.DeregisterTargets(input)

		if err != nil {
			panic(err)
		}

		fmt.Println(input)
	}
}

func (this DeregisterOldInstancesAction) Rollback(pipelineInfo *PipelineInfo) {
	targets := []*elbv2.TargetDescription{}

	for _, instanceId := range pipelineInfo.OldInstancesIds {
		targets = append(targets, &elbv2.TargetDescription{Id: instanceId})
	}

	for _, tgArn := range pipelineInfo.TargetGroupsArns {
		input := &elbv2.RegisterTargetsInput{
			TargetGroupArn: tgArn,
			Targets: targets,
		}
		
		_, err := this.Svc.RegisterTargets(input)

		if err != nil {
			panic(err)
		}
	}
}

func (this WaitForDeregisterAction) Commit(pipelineInfo *PipelineInfo) {
	targets := []*elbv2.TargetDescription{}

	for _, instanceId := range pipelineInfo.OldInstancesIds {
		targets = append(targets, &elbv2.TargetDescription{Id: instanceId})
	}

	for _, tgArn := range pipelineInfo.TargetGroupsArns {
		input := &elbv2.DescribeTargetHealthInput{
			TargetGroupArn: tgArn,
			Targets: targets,
		}
		
		err := this.Svc.WaitUntilTargetDeregistered(input)

		if err != nil {
			panic(err)
		}
	}
}

func (this WaitForDeregisterAction) Rollback(pipelineInfo *PipelineInfo) {
	
}

func (this TerminateOldInstancesAction) Commit(pipelineInfo *PipelineInfo) {
	input := &ec2.TerminateInstancesInput{InstanceIds: pipelineInfo.OldInstancesIds}

	_, err := this.Svc.TerminateInstances(input)

	if err != nil {
		panic(err)
	}
}

func (this TerminateOldInstancesAction) Rollback(pipelineInfo *PipelineInfo) {
	
}
