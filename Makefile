prebuild:

test: clean  c_smnrpc_autocode 
	smnrpc-autocode -cfg ./datas/cfgs/testrpc.json
	sudo env PATH="$(PATH)" go run ./test/smnrpc/test.go

clean:
	rm -rf ./pbr
	rm -rf ./datas/proto/temp

c_smnrpc_autocode: #SureMoonRPC code tool.
	cd ./cmd/smnrpc-autocode && go install

c_smwget: #check md5sum before call wget.
	cd ./cmd/smwget && go install

c_gogopb: #change pb to google.golang ver.
	cd ./cmd/gogopb && go install 

c_smake:
	cd ./cmd/smake && go install 

c_smgit:
	cd ./cmd/smgit && go install

c_smdcatalog:
	cd ./cmd/smdcatalog && go install

c_smkdir:
	cd ./cmd/smkdir && go install 

install: auto_code  c_smcfg  c_smnrpc_autocode c_smwget  c_gogopb c_smake c_smdcatalog c_smgit c_smkdir
	 smdcatalog 

c_smcfg: # a config tool
	cd ./cmd/smcfg && go install 

qrun: c_smnrpc_autocode 
	smnrpc-autocode

auto_code: nothing
	go run build.go 

nothing:
