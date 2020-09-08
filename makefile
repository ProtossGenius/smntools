test: clean  c_smnrpc_autocode 
	smnrpc-autocode -cfg ./datas/cfgs/testrpc.json
	mv ./go/pb ./pb
	rm -rf ./go
	time -p go run ./test/smnrpc/test.go

clean:
	rm -rf ./pbr
	rm -rf ./rpc_nitf
	rm -f ./datas/proto/rip_*.proto ./datas/proto/smn_dict.proto
	rm -rf ./datas/proto/temp
	rm -rf ./pb

c_smnrpc_autocode: #SureMoonRPC code tool.
	cd ./cmd/smnrpc-autocode && go install

c_smwget: #check md5sum before call wget.
	cd ./cmd/smwget && go install

c_gogopb: #change pb to google.golang ver.
	cd ./cmd/gogopb && go install 

c_smake:
	cd ./cmd/smake && go install 

install: auto_code  c_smcfg  c_smnrpc_autocode c_smwget  c_gogopb c_smake

c_smcfg: # a config tool
	cd ./cmd/smcfg && go install 

qrun: c_smnrpc_autocode 
	smnrpc-autocode

auto_code: nothing
	go run build.go 

nothing:
