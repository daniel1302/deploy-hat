
package main

import(
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

    "errors"
)

func isIpAuthorized(svc *ec2.EC2, sgId string, port int64, ip string) (bool, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{
			aws.String(sgId),
		},
	}

	result, err := svc.DescribeSecurityGroups(input)
	if err != nil {
        panic(err)
    }

	if len(result.SecurityGroups) < 1 {
        return false, errors.New("Security group does not exists.")
    }

    ipPermissions := result.SecurityGroups[0].IpPermissions

    for _, item := range ipPermissions {
        if *item.FromPort > port || *item.ToPort < port {
            continue
        }

        for _, ipPerm := range item.IpRanges {
            if *ipPerm.CidrIp == string(ip + "/32") {
                return true, nil
            }
        }
    }


    return false, nil
}

func authorizeIp(svc *ec2.EC2, sgId string, port int64, ip string) error {
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(sgId),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(port),
				IpProtocol: aws.String("tcp"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String(ip + "/32"),
						Description: aws.String("SSH access from the LA office"),
					},
				},
				ToPort: aws.Int64(port),
			},
		},
	}

	_, err := svc.AuthorizeSecurityGroupIngress(input)
	if err != nil {
		return err
	}

	return nil
}

func revokeIp(svc *ec2.EC2, sgId string, port int64, ip string) error {

	input := &ec2.RevokeSecurityGroupIngressInput{
		GroupId: aws.String(sgId),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(port),
				IpProtocol: aws.String("tcp"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String(ip + "/32"),
						Description: aws.String("SSH access from the LA office"),
					},
				},
				ToPort: aws.Int64(port),
			},
		 },
	}

	_, err := svc.RevokeSecurityGroupIngress(input)
	if err != nil {
		return err
	}

	return nil
}