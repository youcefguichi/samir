# TODO: move each alias to the adequate app folder
generate_ca:
	go build -o bin/certs_generator certs_generator/certs_generator.go certs_generator/main.go
	./bin/certs_generator --ca ca.crt ca.key
generate_self_signed_cert:
	go build -o bin/certs_generator certs_generator/certs_generator.go certs_generator/main.go
	./bin/certs_generator --ca-crt-location certs/ca.crt --ca-key-location certs/ca.key $(for) $(crt) $(key)
run_application_load_balancer:
	go build -o bin/application_loadbalancer app_load_balancer/application_loadbalancer.go app_load_balancer/main.go
	./bin/application_loadbalancer
run_tcp_proxy:
	go build -o bin/tcp_proxy tcp_proxy/main.go
	./bin/tcp_proxy
run_backend_servers:
	python3 -m http.server 8080 &
	python3 -m http.server 8081 &
run_client:
	go build -o bin/client client/main.go
	./bin/client