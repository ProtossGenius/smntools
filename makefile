test: clean  c_smnrpc_autocode 
	smnrpc-autocode -cfg ./datas/cfgs/testrpc.json
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

c_smnrpc_autocode:
	cd ./cmd/smnrpc-autocode && go install

c_smwget:
	cd ./cmd/smwget && go install

install: auto_code  c_smcfg c_smpf c_asppl c_smnrpc_autocode c_smwget

c_smcfg:
	cd ./cmd/smcfg && go install 

qrun: c_smnrpc_autocode 
	smnrpc-autocode

auto_code: nothing
	go run build.go 

nothing:
