export CLONED_NAMESPACE_PARENT_PID=404491

# setup bridge with vnet 10.0.0.0/24, and assign 10.0.0.1 to the bridge.
sudo ip link add name samir0 type bridge
sudo ip addr add 10.0.0.1/24 dev samir0
sudo ip link set samir0 up


# create a link between two virtual interfaces,
# just like any ethernet caple between two interfaces. 
sudo ip link add veth0 type veth peer name veth0-cont

sudo ip link set veth0 up

# attach veth0 to samir0, from the host side
# example fom my host machine:
# veth0@if24: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master samir0 state UP mode DEFAULT group default qlen 1000
# link/ether xx:xx:xx:xx:xx:xx brd xx:xx:xx:xx:xx:xx link-netnsid 0
sudo ip link set veth0 master samir0 

# attach vnet0-cont to the cloned namespace parent PID.
sudo ip link set veth0-cont netns $CLONED_NAMESPACE_PARENT_PID

# start veth0-cont inside the namespace
# attach an ip adress form the bridge network (samir0) (this later changed to use dhcp)
# add a default route for the namespace that all trafic should be forwarded to the bridge ip.
sudo nsenter -t $CLONED_NAMESPACE_PARENT_PID -n ip link set veth0-cont up
sudo nsenter -t $CLONED_NAMESPACE_PARENT_PID -n ip addr add 10.0.0.2/24 dev veth0-cont # [1]
sudo nsenter -t $CLONED_NAMESPACE_PARENT_PID -n ip route add default via 10.0.0.1

# to access internet we would need to snat to our host machine
# this operation done by -j MASQUERADE. which works as NAT.
iptables -t nat -A POSTROUTING -s 10.0.0.0/24 ! -o samir0 -j MASQUERADE

# in order for the packets to travel from the bridge network 
# to the outside world, we would need to enable IP forwarding
# so the host doesn't drop the packets that is not destined to the host, which is the default.
sudo sysctl -w net.ipv4.ip_forward=1


# DHCP
# in order to to not specify statically the address ip we did above on [1].
# i setup a lightwight dchp server that runs on the bridge network.

sudo apt install dnsmasq

# the dhcp server will give ip addresses to the parent pids 
# running inside the namespace
sudo cat <<EOF >> /etc/dnsmasq.conf
port=1015
interface=samir0
dhcp-range=10.0.0.10,10.0.0.100,255.255.255.0,12h
EOF

systemctl restart dnsmasq

# one last step is the parent namespace runningin the cloned namespaces
# would need to run a dhcp client that request an ip from the dhcp server.

sudo nsenter -t $CLONED_NAMESPACE_PARENT_PID -n udhcpc -i veth0-cont

# example output
# udhcpc: started, v1.36.0
# udhcpc: broadcasting discover
# udhcpc: broadcasting discover
# udhcpc: broadcasting select for 10.0.0.34, server 10.0.0.1
# udhcpc: lease of 10.0.0.34 obtained from 10.0.0.1, lease time 43200
