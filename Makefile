go-dependencies:
	# https://asdf-vm.com/
	asdf install golang || :

	# https://github.com/securego/gosec
	go install github.com/securego/gosec/v2/cmd/gosec@latest

	go install github.com/nunnatsa/ginkgolinter/cmd/ginkgolinter@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

	asdf reshim golang || :

	go get -u -t -v ./... || :

go-all: go-dependencies
	go env
	go generate ./...
	go mod tidy || :
	gosec ./...
	golangci-lint run ./...
	go test ./...
