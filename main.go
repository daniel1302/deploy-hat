package main

import (
	_ "github.com/aws/aws-sdk-go/aws/session"
	_ "github.com/aws/aws-sdk-go/service/ec2"

	"fmt"
	"os"
)

// InputArgs is struct to keep input arguments
type InputArgs struct {
	OldAMI string
	NewAMI string
}

type RecoveryAction interface {
    Recovery()
}

type ApplyAction interface {
	Apply()
}



func softExit(exitStatus int, message string) {
	fmt.Println(message)
	os.Exit(exitStatus)
}

func getInstancesId(oldAmiID string) {

}

func launchNewInstance(newAmiID string) {

}

func waitUntilReady(instanceID string) {

}

func killInstance(instanceID string) {

}

func detachInstance(instanceId string, asgName string) {

}

func attachInstance(instance string asgName string) {

}

func checkNewAMI(newAmiID string) {

}



func main() {
	if len(os.Args) != 3 {
		softExit(1, fmt.Sprintf("Invalid usage. Usage: %s OLD_AMI_ID, NEW_AMI_ID", os.Args[0]))
	}

	inputArgs := InputArgs{os.Args[1], os.Args[2]}
}