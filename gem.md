https://gist.github.com/denji/12b3a568f092ab951456
- https://github.com/aaydin-tr/divisor/blob/6434d95eea947a11e7f6c7bea8417660a23412d1/internal/proxy/proxy_test.go#L188




- (Layer 4) loadBalancerIP: 20.1.1.1
- (Layer 7) loadBalancerNginx: 10.1.1.1


*.test.com
20.1.1.1





youcef.test.com  > 20.1.1.1 > 10.1.1.1 > check tls and establish a secure connection, check host, if youcef.test.com then forwards the traffic to service-a

