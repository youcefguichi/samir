package main


func main() {
	lb := lb{}
	lb.LoadConfig("config.yaml")
	lb.start()
}
