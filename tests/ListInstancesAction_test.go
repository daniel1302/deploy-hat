package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/stretchr/testify/assert"
	"errors"
)

type mockEC2ClientError struct {
    ec2iface.EC2API
}

func (t mockEC2ClientError) DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return nil, errors.New("Error_From_Ec2Client")
}

type mockEC2ClientCorrectResult struct {
    ec2iface.EC2API
}

func (t mockEC2ClientCorrectResult) DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	res := &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					{
						InstanceId: aws.String("i-1"),
						InstanceType: aws.String("t2-micro"),
						KeyName: aws.String("prod-bastion-key"),
						SubnetId: aws.String("subnet-123a"),
						VpcId: aws.String("vpc-prod"),
						SecurityGroups: []*ec2.GroupIdentifier{ { GroupId: aws.String("sg-123") }, { GroupId: aws.String("sg-567") } },
						Tags: []*ec2.Tag{
							{ Key: aws.String("Name"), Value: aws.String("web-application-1") },
							{ Key: aws.String("Version"), Value: aws.String("123") },
						},
					},
				},
			},
			{
				Instances: []*ec2.Instance{
					{
						InstanceId: aws.String("i-2"),
						InstanceType: aws.String("t2-micro"),
						KeyName: aws.String("prod-bastion-key"),
						SubnetId: aws.String("subnet-123b"),
						VpcId: aws.String("vpc-prod"),
						SecurityGroups: []*ec2.GroupIdentifier{ { GroupId: aws.String("sg-123") }, { GroupId: aws.String("sg-898") } },
						Tags: []*ec2.Tag{
							{ Key: aws.String("Name"), Value: aws.String("web-application-2") },
							{ Key: aws.String("Version"), Value: aws.String("123") },
						},
					},
				},
			},
		},
	}

	return res, nil
}

type mockEC2ClientNoResult struct {
    ec2iface.EC2API
}

func (t mockEC2ClientNoResult) DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	res := &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{},
	}

	return res, nil
}

func TestInitializePipelineActionErrorFromClient(t *testing.T) {
	pipelineInfo := &PipelineInfo{}
	svcMock := &mockEC2ClientError{}

	err := ListInstancesAction{svcMock}.Commit(pipelineInfo)

	if err == nil || err.Error() != "Error_From_Ec2Client" {
		t.Error("Expected error from ListInstancesAction.Commit()")
	}
}

func TestInitializePipelineActionDescribeInstances(t *testing.T) {
	pipelineInfo := &PipelineInfo{}
	svcMock := &mockEC2ClientCorrectResult{}

	ListInstancesAction{svcMock}.Commit(pipelineInfo)

	expectedPipelineInfo := PipelineInfo{
		OldInstancesIds: []*string{},
		OldInstances: []ShortInstanceDesc{
			{
				Id: "i-1",
				InstanceType: "t2-micro",
				KeyName: "prod-bastion-key",
				SubnetId: "subnet-123a",
				VpcId: "vpc-prod",
				SecurityGroupsIds: []*string{ aws.String("sg-123"), aws.String("sg-567") },
				Tags: map[string]string{ "Name": "web-application-1", "Version": "123" },
			},
			{
				Id: "i-2",
				InstanceType: "t2-micro",
				KeyName: "prod-bastion-key",
				SubnetId: "subnet-123b",
				VpcId: "vpc-prod",
				SecurityGroupsIds:  []*string{ aws.String("sg-123"), aws.String("sg-898") },
				Tags: map[string]string{ "Name": "web-application-2", "Version": "123" },
			},
		},
	}

	if len(pipelineInfo.OldInstances) != len(expectedPipelineInfo.OldInstances) {
		t.Errorf("Invalid result from ListInstancesAction.Commit(). Expected %d instances. Got %d.", len(expectedPipelineInfo.OldInstances), len(pipelineInfo.OldInstances))
	}

	for idx, instance := range pipelineInfo.OldInstances {
		assert.Equal(t, instance.Id, expectedPipelineInfo.OldInstances[idx].Id)
		assert.Equal(t, instance.InstanceType, expectedPipelineInfo.OldInstances[idx].InstanceType)
		assert.Equal(t, instance.KeyName, expectedPipelineInfo.OldInstances[idx].KeyName)
		assert.Equal(t, instance.SubnetId, expectedPipelineInfo.OldInstances[idx].SubnetId)
		assert.Equal(t, instance.VpcId, expectedPipelineInfo.OldInstances[idx].VpcId)
		assert.Equal(t, instance.SecurityGroupsIds, expectedPipelineInfo.OldInstances[idx].SecurityGroupsIds)
		assert.Equal(t, instance.Tags, expectedPipelineInfo.OldInstances[idx].Tags)
	}
}

func TestInitializePipelineActionNoInstances(t *testing.T) {
	pipelineInfo := &PipelineInfo{}
	svcMock := &mockEC2ClientNoResult{}

	err :=ListInstancesAction{svcMock}.Commit(pipelineInfo)
	if err == nil {
		t.Error("Expected error from ListInstancesAction.Commit()")
	}
	assert.Equal(t, "Not found any running instance", err.Error())
}
