package main

func main() {

	CheckCertsforAllPeers([]string{"localhost:3002","localhost:3000", "crashloop.sh:443", "biznesbees.com:443"}, true)
}
