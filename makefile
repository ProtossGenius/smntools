
test: nothing
	smn_itf2proto -i "./smnitf" -o ./datas/proto
	smn_protocpl -i ./datas/proto/  -o "./pb/" -ep "github.com/ProtossGenius/smntools"
	smn_pr_go -proto "./datas/proto/" -pkgh "pb/" -o "./pbr/read.go" -gopath=$(GOPATH)/src -ext="/github.com/ProtossGenius/smntools"
	smn_itf2rpc_go -i "./smnitf/" -s -c -o "./rpc_nitf" -gopath=$(GOPATH)/src -pkgh="github.com/ProtossGenius"
	go run ./test/smnrpc/test.go
clean:
	rm -rf ./pbr
	rm -rf ./rpc_nitf
	rm -f ./datas/proto/rip_*.proto ./datas/proto/smn_dict.proto
	rm -rf ./pb

c_smgoget:
	cd ./cmd/smgoget/ && go install

install: c_smgoget


import:
	go get -u github.com/json-iterator/go
	go get -u github.com/ProtossGenius/SureMoonNet
	cd $(GOPATH)/src/github.com/ProtossGenius/SureMoonNet && make install

nothing: