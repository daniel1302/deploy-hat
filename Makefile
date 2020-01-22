build : dependencies lint compile
	
compile :
	go build -o deploy main.go 
	
run :
	go run main.go

test : lint compile
	./deploy --config=./config.yml.dict

lint :
	golint -set_exit_status main.go

dependencies :
	go get -u golang.org/x/lint/golint
	go get -u github.com/awslabs/aws-sdk-go/aws
	go get -u github.com/aws/aws-sdk-go/aws/session
	go get -u github.com/aws/aws-sdk-go/service/ec2