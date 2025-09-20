WAN=$(ip route | awk '/^default/ {print $5; exit}'); echo "WAN=$WAN"
sudo sysctl -w net.ipv4.ip_forward=1
sudo sysctl -w net.ipv4.conf.all.rp_filter=0
sudo sysctl -w net.ipv4.conf.default.rp_filter=0
sudo sysctl -w net.ipv4.conf.samir-br.rp_filter=0
sudo sysctl -w net.ipv4.conf."$WAN".rp_filter=0
sudo iptables -C FORWARD -i samir-br -o "$WAN" -j ACCEPT 2>/dev/null || \
sudo iptables -I FORWARD 1 -i samir-br -o "$WAN" -j ACCEPT
sudo iptables -C FORWARD -i "$WAN" -o samir-br -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || \
sudo iptables -I FORWARD 1 -i "$WAN" -o samir-br -m state --state RELATED,ESTABLISHED -j ACCEPT
sudo iptables -t nat -C POSTROUTING -s 10.10.0.0/16 -o "$WAN" -j MASQUERADE 2>/dev/null || \
sudo iptables -t nat -A POSTROUTING -s 10.10.0.0/16 -o "$WAN" -j MASQUERADE
