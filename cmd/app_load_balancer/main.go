package main

func main() {
	app_loadbalancer := lb{}
	app_loadbalancer.LoadConfig("/app_loadbalancer_config.yaml")
	app_loadbalancer.start()
}