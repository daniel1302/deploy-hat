package main

import (
	"testing"
)

func TestInitializePipelineAction(t *testing.T) {

	dataTable := []struct{
		OldAMI string
		NewAMI string
		expectedError bool
	}{
		{"ami-123", "ami-213", false},
		{"ami-1ds1ha9822123", "ami-7dsa7fg7r7b7q1", false},
		{"ami-123", "ami-123", true},
		{"ami-123", "", true},
		{"", "ami-123", true},
	}

	for _, item := range dataTable {
		pipelineInfo := &PipelineInfo{}
		action := InitializePipelineAction{item.OldAMI, item.NewAMI}
		err := action.Commit(pipelineInfo)

		if item.expectedError == true {
			if  err == nil {
				t.Errorf("Function InitializePipelineAction.Commit(%v) executed correctly. Expected error.", item)
			}

			continue;
		}

		if err != nil {
			t.Errorf("Unexpected error for InitializePipelineAction.Commit(%v) method. Error: %s", item, err.Error())
			continue
		}

		if item.OldAMI != pipelineInfo.Input.OldAMI {
			t.Errorf("Old AMI ID is valid in pipeline. Expected %s. Got %s.", item.OldAMI, pipelineInfo.Input.OldAMI)
		}

		if item.NewAMI != pipelineInfo.Input.NewAMI {
			t.Errorf("New AMI ID is valid in pipeline. Expected %s. Got %s.", item.NewAMI, pipelineInfo.Input.NewAMI)
		}

		if action.Rollback(pipelineInfo) != nil {
			t.Errorf("Rollback failed for InitializePipelineAction.")
		}
	}
}
