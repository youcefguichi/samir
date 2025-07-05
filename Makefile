generate_ca:
	go build -o bin/certs_generator cmd/certs_generator/certs_generator.go cmd/certs_generator/main.go
	./bin/certs_generator --ca ca.crt ca.key
generate_self_signed_cert:
	go build -o bin/certs_generator cmd/certs_generator/certs_generator.go cmd/certs_generator/main.go
	./bin/certs_generator --ca-crt-location certs/ca.crt --ca-key-location certs/ca.key $(for) $(crt) $(key)
run_application_load_balancer:
	go build -o bin/application_loadbalancer cmd/app_load_balancer/application_loadbalancer.go cmd/app_load_balancer/main.go
	./bin/application_loadbalancer
run_tcp_proxy:
	go build -o bin/tcp_proxy cmd/tcp_proxy/main.go
	./bin/tcp_proxy
run_backend_servers:
	python3 -m http.server 8080 &
	python3 -m http.server 8081 &
run_client:
	go build -o bin/client cmd/client/main.go
	./bin/client