run_container:
	go build -o container container.go main.go
	sudo ./container run /bin/sh
move_samir_to_bin:
	go build -o samir container.go main.go
	sudo mv ./samir /usr/local/bin/samir
	sudo chmod +x /usr/local/bin/samir
