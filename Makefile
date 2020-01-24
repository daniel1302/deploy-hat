SOURCES = $(shell find src/ ! -name "*_test.go" -name "*.go")
TESTS_SRC = $(wildcard src/*_test.go)

build : dependencies lint compile
	
compile :
	go build -o deploy $(SOURCES)
	
run :
	go run $(SOURCES)

test : compile
	go test -v $(SOURCES) $(TESTS_SRC)

lint :
	golint -set_exit_status $(SOURCES)

dependencies :
	go get -u golang.org/x/lint/golint
	go get -u github.com/aws/aws-sdk-go/aws
	go get -u github.com/aws/aws-sdk-go/aws/session
	go get -u github.com/aws/aws-sdk-go/service/ec2
	go get -u github.com/aws/aws-sdk-go/service/elbv2
	go get -u github.com/aws/aws-sdk-go/service/ec2/ec2iface
	go get -u github.com/stretchr/testify/assert
