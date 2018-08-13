build:
	GOOS=linux GOARCH=amd64 go build -o godisk *.go

deploy:
	GOOS=linux GOARCH=amd64 go build -o godisk *.go
	tar czvf godisk.tar godisk 
	ssh root@vm2 "mkdir /root/godisk"
	scp godisk.tar root@vm2:/root/godisk
	ssh root@vm2 "cd /root/godisk && tar zxvf godisk.tar && rm -rf godisk.tar"
	rm -rf godisk.tar
