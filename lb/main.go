package main

func main() {
	lb := NewLoadBalancer("localhost", "5556", "https", "server.crt", "server.key", []string{"backend1", "backend2"})
	lb.start()
	lb.getNextBackend()
}
