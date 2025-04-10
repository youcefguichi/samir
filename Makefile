jumpHost:
	go build -o jumpHost jumpHost.go
	./jumpHost
	rm jumpHost
	go clean
layer7:
	go build -o layer7 layer7.go
	./layer7
	rm layer7
	go clean
client:
	go build -o client client.go
	./client
	rm client
	go clean