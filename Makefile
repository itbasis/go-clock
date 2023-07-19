go-install:
	# https://asdf-vm.com/
	asdf install golang || :

	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/nunnatsa/ginkgolinter/cmd/ginkgolinter@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

	asdf reshim golang || :

	go get -u -t -v ./... || :

go-test: go-install
	go env
	gosec ./...
	golangci-lint run ./...
	go test ./...
