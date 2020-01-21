package main

import (
	_ "github.com/aws/aws-sdk-go/aws/session"
	_ "github.com/aws/aws-sdk-go/service/ec2"

	"fmt"
	"os"
	"errors"
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



// func main() {
// 	if len(os.Args) != 3 {
// 		softExit(1, fmt.Sprintf("Invalid usage. Usage: %s OLD_AMI_ID, NEW_AMI_ID", os.Args[0]))
// 	}

// 	inputArgs := InputArgs{os.Args[1], os.Args[2]}
// }


type InfrastructureAction interface {
    Apply()
    Recovery()
}

type RecoveryStack []InfrastructureAction

type Step1 struct {
    id string
}

type Step2 struct {
    id string
}

func (s Step1) Apply() {
    fmt.Println("Step1 apply", s.id)
}


func (s Step1) Recovery() {
    fmt.Println("Step1 recovery", s.id)
}

func (s Step2) Apply() {
    fmt.Println("Step2 apply", s.id)
}


func (s Step2) Recovery() {
    fmt.Println("Step2 recovery", s.id)
}


func Perform() (err error) {
   defer func() {
      if r := recover(); r != nil {
         err = r.(error)
      }
	
   }()
   GoesWrong()
   return
}

func GoesWrong() {
   panic(errors.New("Fail"))
}

func main() {
   var s1 InfrastructureAction = &Step1{"some_unique_id_for_step1"}
   var s2 InfrastructureAction = &Step2{"some_unique_id_for_step2"}

   var rStack RecoveryStack

   rStack = append(rStack, s1)
   rStack = append(rStack, s2)

   for idx, _ := range rStack {
	rStack[idx].Apply()
   }
}