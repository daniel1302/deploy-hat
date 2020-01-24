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

// InfrastructureAction is an interface for all deployment steps
type InfrastructureAction interface {
    Commit(info *PipelineInfo) error
    Rollback(info *PipelineInfo) error
}

// ShortInstanceDesc keeps information about instance required for the deployment
type ShortInstanceDesc struct {
    ID string
    InstanceType string
    KeyName string
    SubnetID string
    VpcID string
    SecurityGroupsIds []*string
    Tags map[string]string
}

// PipelineInfo keep information about deployment progress
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

// InitializePipelineAction is a pipeline step struct
type InitializePipelineAction struct {
    OldAMI string
    NewAMI string
}

// ListInstancesAction is a pipeline step struct
type ListInstancesAction struct {
    Svc   ec2iface.EC2API
}

// RunInstancesAction is a pipeline step struct
type RunInstancesAction struct {
    Svc   *ec2.EC2
}

// WaitUntilStatusOkAction is a pipeline step struct
type WaitUntilStatusOkAction struct {
    Svc   *ec2.EC2
}

// AuthorizeSecurityGroupsAction is a pipeline step struct
type AuthorizeSecurityGroupsAction struct {
    Svc   *ec2.EC2
}

// TestInstancesAction is a pipeline step struct
type TestInstancesAction struct {
    Svc   *ec2.EC2
}

// CollectPublicIpsAction is a pipeline step struct
type CollectPublicIpsAction struct {
    Svc   *ec2.EC2
}

// FindLoadBalancerAction is a pipeline step struct
type FindLoadBalancerAction struct {
    Svc   *elbv2.ELBV2
}

// RegisterNewInstancesAction is a pipeline step struct
type RegisterNewInstancesAction struct {
    Svc   *elbv2.ELBV2
}

// WaitForDeregisterAction is a pipeline step struct
type WaitForDeregisterAction struct {
    Svc   *elbv2.ELBV2
}

// DeregisterOldInstancesAction is a pipeline step struct
type DeregisterOldInstancesAction struct {
    Svc   *elbv2.ELBV2
}

// TerminateOldInstancesAction is a pipeline step struct
type TerminateOldInstancesAction struct {
    Svc   *ec2.EC2
}

// Commit is an action to apply changes in the InitializePipelineAction step
func (act InitializePipelineAction) Commit(pipelineInfo *PipelineInfo) error {
    OldAMI := act.OldAMI
    NewAMI := act.NewAMI

    if len(OldAMI) < 4|| len(NewAMI) < 4 {
        return errors.New("Invalid AMI ID")
    }

    if OldAMI[:4] != "ami-" || NewAMI[:4] != "ami-" {
        return errors.New("Invalid AMI ID")
    }

    if OldAMI == NewAMI {
        return errors.New("Both new and old AMI ID are the same")
    }

    pipelineInfo.Input = InputArgs{OldAMI, NewAMI}
    clientIP, err := getClientIP()

    if err != nil {
        return err
    }

    pipelineInfo.ClientIP = clientIP
    return nil
}

// Rollback is an action to apply changes in the InitializePipelineAction step
func (act InitializePipelineAction) Rollback(pipelineInfo *PipelineInfo) error {
    return nil
}

// Commit is an action to apply changes in the ListInstancesAction step
func (act ListInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
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

    result, err := act.Svc.DescribeInstances(input)
    if err != nil {
        return err
    }

    for _, item := range result.Reservations {
        instance := item.Instances[0]

        sgIDs := []*string{}

        for _, sg := range instance.SecurityGroups {
            sgIDs = append(sgIDs, sg.GroupId)
        }

        tags := map[string]string{}

        for _, tag := range instance.Tags {
            tags[*tag.Key] = *tag.Value
        }

        pipelineInfo.OldInstancesIds = append(pipelineInfo.OldInstancesIds, instance.InstanceId)
        pipelineInfo.OldInstances = append(pipelineInfo.OldInstances, ShortInstanceDesc{
            ID: *instance.InstanceId,
            InstanceType: *instance.InstanceType,
            KeyName: *instance.KeyName,
            SubnetID: *instance.SubnetId,
            VpcID: *instance.VpcId,
            SecurityGroupsIds: sgIDs,
            Tags: tags,
        })
    }

    if len(pipelineInfo.OldInstances) < 1 {
        return errors.New("Not found any running instance")
    }

    return nil
}

// Rollback is an action to apply changes in the ListInstancesAction step
func (act ListInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
    return nil
}

// Commit is an action to apply changes in the RunInstancesAction step
func (act RunInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
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
                    SubnetId:                 aws.String(item.SubnetID),
                    Groups:                   item.SecurityGroupsIds,
                },
            },
            TagSpecifications: []*ec2.TagSpecification{
                {
                    ResourceType: aws.String("instance"),
                    Tags: newTags,
                },
            },
        }

        result, err := act.Svc.RunInstances(input)
        if err != nil {
            return err
        }

        pipelineInfo.NewInstancesIds = append(pipelineInfo.NewInstancesIds, result.Instances[0].InstanceId)
    }

    return nil
}

// Rollback is an action to apply changes in the RunInstancesAction step
func (act RunInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
    input := &ec2.TerminateInstancesInput{
        InstanceIds: pipelineInfo.NewInstancesIds,
    }

    _, err := act.Svc.TerminateInstances(input)
    if err != nil {
        return err
    }

    return nil
}

// Commit is an action to apply changes in the WaitUntilStatusOkAction step
func (act WaitUntilStatusOkAction) Commit(pipelineInfo *PipelineInfo) error {

    input := &ec2.DescribeInstancesInput{
        InstanceIds: pipelineInfo.NewInstancesIds,
	}

    err := act.Svc.WaitUntilInstanceRunning(input)

    if err != nil {
        return err
    }

    return nil
}

// Rollback is an action to apply changes in the WaitUntilStatusOkAction step
func (act WaitUntilStatusOkAction) Rollback(pipelineInfo *PipelineInfo) error {
    return nil
}

// Commit is an action to apply changes in the AuthorizeSecurityGroupsAction step
func (act AuthorizeSecurityGroupsAction) Commit(pipelineInfo *PipelineInfo) error {

    for _, instance := range pipelineInfo.OldInstances {
        if len(instance.SecurityGroupsIds) < 1 {
            return errors.New("Instance must have Security Group")
        }

        isIPAuthorizedStatus, err := isIPAuthorized(act.Svc, *instance.SecurityGroupsIds[0], 80, pipelineInfo.ClientIP)
        if err != nil {
            return errors.New("Cannot describe security group")
        }

        if !isIPAuthorizedStatus {
            authorizeIP(act.Svc, *instance.SecurityGroupsIds[0], 80, pipelineInfo.ClientIP)
            pipelineInfo.ModifiedSecurityGroups = append(pipelineInfo.ModifiedSecurityGroups, instance.SecurityGroupsIds[0])
        }
    }

    return nil
}

// Rollback is an action to apply changes in the AuthorizeSecurityGroupsAction step
func (act AuthorizeSecurityGroupsAction) Rollback(pipelineInfo *PipelineInfo) error {
    for _, sgID := range pipelineInfo.ModifiedSecurityGroups {
        err := revokeIP(act.Svc, *sgID, 80, pipelineInfo.ClientIP)

        if err != nil {
            return errors.New("Cannot describe security group")
        }
    }

    return nil
}

// Commit is an action to apply changes in the CollectPublicIpsAction step
func (act CollectPublicIpsAction) Commit(pipelineInfo *PipelineInfo) error {

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

    result, err := act.Svc.DescribeInstances(input)
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

// Rollback is an action to apply changes in the CollectPublicIpsAction step
func (act CollectPublicIpsAction) Rollback(pipelineInfo *PipelineInfo) error {
	return nil
}

// Commit is an action to apply changes in the TestInstancesAction step
func (act TestInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
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

// Rollback is an action to apply changes in the TestInstancesAction step
func (act TestInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
	return nil
}

// Commit is an action to apply changes in the FindLoadBalancerAction step
func (act FindLoadBalancerAction) Commit(pipelineInfo *PipelineInfo) error {
    input := &elbv2.DescribeTargetGroupsInput{
        PageSize: aws.Int64(400),
    }

    res, err := act.Svc.DescribeTargetGroups(input)

    if err != nil {
        return err
    }

    for _, tg := range res.TargetGroups {
        status, err := findInstancesInTargetGroup(act.Svc, *tg.TargetGroupArn, pipelineInfo.OldInstancesIds)

        if err != nil {
            return err
        }
        if status == true {
            pipelineInfo.TargetGroupsArns = append(pipelineInfo.TargetGroupsArns, tg.TargetGroupArn)
        }
	}

	return nil
}

// Rollback is an action to apply changes in the FindLoadBalancerAction step
func (act FindLoadBalancerAction) Rollback(pipelineInfo *PipelineInfo) error {
	return nil
}

// Commit is an action to apply changes in the RegisterNewInstancesAction step
func (act RegisterNewInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
    targets := []*elbv2.TargetDescription{}

    for _, instanceID := range pipelineInfo.NewInstancesIds {
        targets = append(targets, &elbv2.TargetDescription{Id: instanceID})
    }

    for _, tgArn := range pipelineInfo.TargetGroupsArns {
        input := &elbv2.RegisterTargetsInput{
            TargetGroupArn: tgArn,
            Targets: targets,
        }

        _, err := act.Svc.RegisterTargets(input)

        if err != nil {
            return err
        }
	}

	return nil
}

// Rollback is an action to apply changes in the RegisterNewInstancesAction step
func (act RegisterNewInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
    targets := []*elbv2.TargetDescription{}

    for _, instanceID := range pipelineInfo.NewInstancesIds {
        targets = append(targets, &elbv2.TargetDescription{Id: instanceID})
    }

    for _, tgArn := range pipelineInfo.TargetGroupsArns {
        input := &elbv2.DeregisterTargetsInput{
            TargetGroupArn: tgArn,
            Targets: targets,
        }

        _, err := act.Svc.DeregisterTargets(input)

        if err != nil {
            return err
        }
	}

	return nil
}

// Commit is an action to apply changes in the DeregisterOldInstancesAction step
func (act DeregisterOldInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
    targets := []*elbv2.TargetDescription{}

    for _, instanceID := range pipelineInfo.OldInstancesIds {
        targets = append(targets, &elbv2.TargetDescription{Id: instanceID})
    }

    for _, tgArn := range pipelineInfo.TargetGroupsArns {
        input := &elbv2.DeregisterTargetsInput{
            TargetGroupArn: tgArn,
            Targets: targets,
        }

        _, err := act.Svc.DeregisterTargets(input)

        if err != nil {
            return err
        }
	}

	return nil
}

// Rollback is an action to apply changes in the DeregisterOldInstancesAction step
func (act DeregisterOldInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
    targets := []*elbv2.TargetDescription{}

    for _, instanceID := range pipelineInfo.OldInstancesIds {
        targets = append(targets, &elbv2.TargetDescription{Id: instanceID})
    }

    for _, tgArn := range pipelineInfo.TargetGroupsArns {
        input := &elbv2.RegisterTargetsInput{
            TargetGroupArn: tgArn,
            Targets: targets,
        }

        _, err := act.Svc.RegisterTargets(input)

        if err != nil {
            return err
        }
	}

	return nil
}

// Commit is an action to apply changes in the WaitForDeregisterAction step
func (act WaitForDeregisterAction) Commit(pipelineInfo *PipelineInfo) error {
    targets := []*elbv2.TargetDescription{}

    for _, instanceID := range pipelineInfo.OldInstancesIds {
        targets = append(targets, &elbv2.TargetDescription{Id: instanceID})
    }

    for _, tgArn := range pipelineInfo.TargetGroupsArns {
        input := &elbv2.DescribeTargetHealthInput{
            TargetGroupArn: tgArn,
            Targets: targets,
        }

        err := act.Svc.WaitUntilTargetDeregistered(input)

        if err != nil {
            return err
        }
	}

	return nil
}

// Rollback is an action to apply changes in the WaitForDeregisterAction step
func (act WaitForDeregisterAction) Rollback(pipelineInfo *PipelineInfo) error {
    return nil
}

// Commit is an action to apply changes in the TerminateOldInstancesAction step
func (act TerminateOldInstancesAction) Commit(pipelineInfo *PipelineInfo) error {
    input := &ec2.TerminateInstancesInput{InstanceIds: pipelineInfo.OldInstancesIds}

    _, err := act.Svc.TerminateInstances(input)

    if err != nil {
        return err
	}

	return nil
}

// Rollback is an action to apply changes in the TerminateOldInstancesAction step
func (act TerminateOldInstancesAction) Rollback(pipelineInfo *PipelineInfo) error {
    return nil
}

