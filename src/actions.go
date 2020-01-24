package main

import(

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/ec2"
    "github.com/aws/aws-sdk-go/service/ec2/ec2iface"
    "github.com/aws/aws-sdk-go/service/elbv2"

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
    Input InputArgs
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
	
    Svc   ec2iface.EC2API
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

    pipelineInfo.Input = InputArgs{OldAMI, NewAMI}
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

func (this ListInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
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
        return err
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
        return errors.New("Not found any running instance")
    }

    return nil
}

func (this ListInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
    return nil
}

func (this RunInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
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
                    Groups:                      item.SecurityGroupsIds,
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
            return err
        }

        pipelineInfo.NewInstancesIds = append(pipelineInfo.NewInstancesIds, result.Instances[0].InstanceId)
    }

    return nil
}

func (this RunInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
    input := &ec2.TerminateInstancesInput{
        InstanceIds: pipelineInfo.NewInstancesIds,
    }

    _, err := this.Svc.TerminateInstances(input)
    if err != nil {
        return err
    }

    return nil
}


func (this WaitUntilStatusOkAction) Commit(pipelineInfo *PipelineInfo) error {

    input := &ec2.DescribeInstancesInput{
        InstanceIds: pipelineInfo.NewInstancesIds,
	}
	
    err := this.Svc.WaitUntilInstanceRunning(input)

    if err != nil {
        return err
    }

    return nil
}

func (this WaitUntilStatusOkAction) Rollback(pipelineInfo *PipelineInfo) error {
    return nil
}

func (this AuthorizeSecurityGroupsAction) Commit(pipelineInfo *PipelineInfo) error {

    for _, instance := range pipelineInfo.OldInstances {
        if len(instance.SecurityGroupsIds) < 1 {
            return errors.New("Instance must have Security Group")
        }

        isIpAuthorizedStatus, err := isIpAuthorized(this.Svc, *instance.SecurityGroupsIds[0], 80, pipelineInfo.ClientIP)
        if err != nil {
            return errors.New("Cannot describe security group")
        }

        if !isIpAuthorizedStatus {
            authorizeIp(this.Svc, *instance.SecurityGroupsIds[0], 80, pipelineInfo.ClientIP)
            pipelineInfo.ModifiedSecurityGroups = append(pipelineInfo.ModifiedSecurityGroups, instance.SecurityGroupsIds[0])
        }
    }

    return nil
}

func (this AuthorizeSecurityGroupsAction) Rollback(pipelineInfo *PipelineInfo) error {
    for _, sgId := range pipelineInfo.ModifiedSecurityGroups {
        err := revokeIp(this.Svc, *sgId, 80, pipelineInfo.ClientIP)

        if err != nil {
            return errors.New("Cannot describe security group")
        }
    }

    return nil
}


func (this CollectPublicIpsAction) Commit(pipelineInfo *PipelineInfo) error {

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
        return err
    }

    for _, item := range result.Reservations {
        instance := item.Instances[0]

        if *instance.PublicIpAddress == "" {
            return errors.New("To perform deployment instance must have public IP")
        }

        pipelineInfo.NewInstancesIps = append(pipelineInfo.NewInstancesIps, *instance.PublicIpAddress)
	}

	return nil
}

func (this CollectPublicIpsAction) Rollback(pipelineInfo *PipelineInfo) error {
	return nil
}



func (this TestInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
    for _, ip := range pipelineInfo.NewInstancesIps {
        isValid, err := isValidRequest("http://"+ip, 10)

        if err != nil {
            return errors.New("Application is down")
        }

        if !isValid {
            return errors.New("Application is down")
        }
	}

	return nil
}

func (this TestInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
	return nil
}




func (this FindLoadBalancerAction) Commit(pipelineInfo *PipelineInfo) error {
    input := &elbv2.DescribeTargetGroupsInput{
        PageSize: aws.Int64(400),
    }

    res, err := this.Svc.DescribeTargetGroups(input)

    if err != nil {
        return err
    }

    for _, tg := range res.TargetGroups {
        status, err := findInstancesInTargetGroup(this.Svc, *tg.TargetGroupArn, pipelineInfo.OldInstancesIds)

        if err != nil {
            return err
        }
        if status == true {
            pipelineInfo.TargetGroupsArns = append(pipelineInfo.TargetGroupsArns, tg.TargetGroupArn)
        }
	}

	return nil
}

func (this FindLoadBalancerAction) Rollback(pipelineInfo *PipelineInfo) error {
	return nil
}


func (this RegisterNewInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
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
            return err
        }
	}

	return nil
}

func (this RegisterNewInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
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
            return err
        }
	}

	return nil
}

func (this DeregisterOldInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
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
            return err
        }
	}

	return nil
}

func (this DeregisterOldInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
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
            return err
        }
	}

	return nil
}

func (this WaitForDeregisterAction) Commit(pipelineInfo *PipelineInfo) error {
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
            return err
        }
	}

	return nil
}

func (this WaitForDeregisterAction) Rollback(pipelineInfo *PipelineInfo) error {
    return nil
}

func (this TerminateOldInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
    input := &ec2.TerminateInstancesInput{InstanceIds: pipelineInfo.OldInstancesIds}

    _, err := this.Svc.TerminateInstances(input)

    if err != nil {
        return err
	}

	return nil
}

func (this TerminateOldInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
    return nil
}
