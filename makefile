test: clean  
	smn_itf2proto -i "./smnitf" -o ./datas/proto
	smn_protocpl -i ./datas/proto/  -o "./pb/" -ep "github.com/ProtossGenius/smntools"
	smn_pr_go -proto "./datas/proto/" -pkgh "pb/" -gopath=$(GOPATH)/src -ext="/github.com/ProtossGenius/smntools"
	smn_itf2rpc_go -i "./smnitf/" -s -c -o "./rpc_nitf" -gopath=$(GOPATH)/src -pkgh="github.com/ProtossGenius"
	go run ./test/smnrpc/test.go
clean:
	rm -rf ./pbr
	rm -rf ./rpc_nitf
	rm -f ./datas/proto/rip_*.proto ./datas/proto/smn_dict.proto
	rm -rf ./datas/proto/temp
	rm -rf ./pb

c_smpf:
	cd ./cmd/smpf && go install 

c_asppl:
	cd ./cmd/asppl && go install 

install: c_smcfg c_smpf c_asppl

c_smcfg:
	cd ./cmd/smcfg && go install 

qrun: c_smcfg
	smcfg -get "git@github.com:ProtossGenius/smcfgs.git"

auto_code: nothing
	go run ./ac_asppl.go

nothing:
