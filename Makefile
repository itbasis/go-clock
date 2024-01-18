go-dependencies:
	# https://asdf-vm.com/
	asdf install golang || :
	#
	# https://github.com/securego/gosec
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
	#
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/nunnatsa/ginkgolinter/cmd/ginkgolinter@latest
	#
	go install github.com/onsi/ginkgo/v2/ginkgo@latest
	#
	#go install github.com/vektra/mockery/v2@latest
	#
	asdf reshim golang || :
	#
	go get -u -t -v ./... || :

go-generate: go-dependencies
	#mockery
	go generate ./...

go-lint: go-dependencies
	#golangci-lint run
	ginkgolinter ./...
	go vet -vettool=$$(go env GOPATH)/bin/shadow ./...

go-test: go-lint
	gosec ./...
	ginkgo -r -race --cover --coverprofile=.coverage-ginkgo.out --junit-report=junit-report.xml ./...
	go tool cover -func=.coverage-ginkgo.out -o=.coverage.out
	cat .coverage.out

go-all-tests: go-dependencies go-generate go-lint go-test

go-all: go-all-tests
	go mod tidy || :
