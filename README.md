### Description

Script used to replace current application instances with the new one.

### Requirements:

- Application AMI has got opened port 80
- Currently working application instances are connected to the Application Load Balancer

### Usage

```
./deploy OLD_AMI NEW_AMI
```

### Example

##### Correct process

```
$ ./deploy ami-0d279985b668e9b38 ami-0aa2563dfc98ff16b

[main.InitializePipelineAction] Executing.
[main.InitializePipelineAction] Finished. No errors
[main.ListInstancesAction] Executing.
[main.ListInstancesAction] Finished. No errors
[main.FindLoadBalancerAction] Executing.
[main.FindLoadBalancerAction] Finished. No errors
[main.RunInstancesAction] Executing.
[main.RunInstancesAction] Finished. No errors
[main.WaitUntilStatusOkAction] Executing.
[main.WaitUntilStatusOkAction] Finished. No errors
[main.AuthorizeSecurityGroupsAction] Executing.
[main.AuthorizeSecurityGroupsAction] Finished. No errors
[main.CollectPublicIpsAction] Executing.
[main.CollectPublicIpsAction] Finished. No errors
[main.TestInstancesAction] Executing.
[main.TestInstancesAction] Finished. No errors
[main.RegisterNewInstancesAction] Executing.
[main.RegisterNewInstancesAction] Finished. No errors
[main.DeregisterOldInstancesAction] Executing.
[main.DeregisterOldInstancesAction] Finished. No errors
[main.WaitForDeregisterAction] Executing.
[main.WaitForDeregisterAction] Finished. No errors
[main.TerminateOldInstancesAction] Executing.
[main.TerminateOldInstancesAction] Finished. No errors
```

##### Process with errors

```
$ ./deploy ami-0d279985b668e9b38 ami-0aa2563dfc98ff16b
[main.InitializePipelineAction] Executing.
[main.InitializePipelineAction] Finished. No errors
[main.ListInstancesAction] Executing.
[main.ListInstancesAction][ERROR] Not found any running instance
[main.ListInstancesAction] Rolling changes back
[main.InitializePipelineAction] Rolling changes back
```



### Development Workflow

##### Prerequisites

- Install GO lang
- Install the make
- Export GOPATH and add $GOPATH/bin to the PATH env variable

##### Compile the project

```
make dependencies
make compile
```

##### Build the project

```
make dependencies
make build
```

##### Check linter errors

```
make dependencies
make lint
```

##### Run unit tests

```
make dependencies
make test
```
