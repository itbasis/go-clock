go-dependencies:
	# https://asdf-vm.com/
	asdf install golang || :

	# https://github.com/securego/gosec
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest

	go install github.com/onsi/ginkgo/v2/ginkgo@latest

	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/nunnatsa/ginkgolinter/cmd/ginkgolinter@latest

	asdf reshim golang || :

	go get -u -t -v ./... || :

go-generate:
	go generate ./...

go-test:
	golangci-lint run ./...
	go vet -vettool=$(which shadow) ./...
	gosec ./...
#	go test ./...
	ginkgo -r -race --cover --coverprofile=.coverage-details.out ./...
	go tool cover -func=.coverage-details.out -o=.coverage.out
	cat .coverage.out

go-all: go-dependencies go-generate go-test
	go mod tidy || :
