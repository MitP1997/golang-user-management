.PHONY: compile-contracts
compile-contracts: generate-abi generate-bin generate-go-files

.PHONY: generate-abi
generate-abi:
	solc --optimize --abi ./contracts/MySmartContract.sol -o build

.PHONY: generate-bin
generate-bin:
	solc --optimize --bin ./contracts/MySmartContract.sol -o build

.PHONY: generate-go-files
generate-go-files:
	abigen --abi=./build/MySmartContract.abi --bin=./build/MySmartContract.bin --pkg=api --out=./api/MySmartContract.go

.PHONY: server
server:
	docker compose up -d
	go run ./cmd/server/main.go

.PHONY: local-dev-env
local-dev-env:
	@scripts/install_codegen_tools.sh
	@scripts/install_brew_packages.sh

.PHONY: gen proto-gen
gen proto-gen:
	buf generate protos --template protos/buf.gen.yaml
	protoc-go-inject-tag -input=./internal/models/*.pb.go -remove_tag_comment
	protoc-go-inject-tag -input=./internal/requests/*.pb.go -remove_tag_comment

.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt format
fmt format:
	go fmt ./...
