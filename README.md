## INFRA FOR FUN WITH GOLANG

Currently the repo have Current services (all services/tools are written from scratch in go):
- a smiple application loadbalancer with roundrobin and simple waf
- a simple tcp proxy
- a simple client with mtls enabled
- a certs generatore for self signed certs
### Use the current services
#### Generate certs for the application load balancer and the clients to enfore mtls
- Genereate a local CA
```
make generate_ca
```
- Generate certs signed by the our local CA that we just created for the clients and for our application loadbalancer
```
make generate_self_signed_cert for=server crt=server.crt key=server.key
make generate_self_signed_cert for=client crt=client.crt key=client.key
```
- Run the application load balancer

```
make run_application_load_balancer
```

the application load balancer will pick the config from the `app_loadbalancer_config.yaml`
```yaml
host: localhost:3002
protocol: https
port: 3002
ca_location: certs/ca.crt
cert_location: certs/server.crt
key_location: certs/server.key
backend_servers:
  - http://localhost:8080
  - http://localhost:8081
waf:
  allowedIPs:
      - 10.0.0.0/8 #0.0.0.0/0 10.0.0.0/8
      - 192.168.1.0/24
      - 172.16.0.0/12
      - 127.0.0.1
  allowedPorts:
      - 80
      - 443
      - "*"
```

you see!! It has some waf also :)

next we we should run our backend servers because as you see it will forwarded the traffic to `localhost:8080` `localhost:8081`

- Run backend servers

  ```
  make run_backend_servers
  ```
as you see the tls will terminate at our application loadbalancer then distribute the traffic to the backends. next step would be to conenct a client to teh application load balancer and watch the the responses from the backends through the load balancer. however before that, i created also a tcp proxy that doesn't understand application layer specifics but only network meaning it only forarad the bytes to the app load balaancer and doesm't understand or inspect the traffic.

- Run the tcp proxy
  ```
  make run_tcp_proxy
  ```
the tcp proxy will pipe the traffic to the application load balancer. next let's connect the client to the tcp proxy. so the traffic will go as follow.
client -> tcp_proxy -> application load balancer -> backend servers

- Run the client
  since we use mtls, the application load balancer will require the client to prove it's identity, that's why we created clients certs previously as well. which they should be signed by teh same CA to prove the identity.
  ```
  make run client
  ```

### Diagram

![Architecture Diagram](./infra-in-go.png)