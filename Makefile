SOURCES = $(wildcard src/*.go)

build : dependencies lint compile
	
compile :
	go build -o deploy $(SOURCES)
	
run :
	go run $(SOURCES)

test : lint compile
	./deploy --config=./config.yml.dict

lint :
	golint -set_exit_status $(SOURCES)

dependencies :
	go get -u golang.org/x/lint/golint
	go get -u github.com/awslabs/aws-sdk-go/aws
	go get -u github.com/aws/aws-sdk-go/aws/session
	go get -u github.com/aws/aws-sdk-go/service/ec2