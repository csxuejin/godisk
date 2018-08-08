build:
	GOOS=linux GOARCH=amd64 go build -o godisk *.go

deploy:
	GOOS=linux GOARCH=amd64 go build -o godisk *.go
	tar czvf godisk.tar godisk 
	scp godisk.tar root@vm3:/root/godisk
	ssh root@vm3 "cd /root/godisk && tar zxvf godisk.tar && rm -rf godisk.tar"
	rm -rf godisk.tar
