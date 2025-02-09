#!/usr/bin/env bash


# this is all LLM trash

# setup wireguard

apt-get update

apt-get install -y wireguard

# setup wireguard
mkdir -p /etc/wireguard

# generate private and public keys
wg genkey | tee /etc/wireguard/privatekey | wg pubkey > /etc/wireguard/publickey

# generate server private and public keys
wg genkey | tee /etc/wireguard/server-privatekey | wg pubkey > /etc/wireguard/server-publickey

# generate client private and public keys
wg genkey | tee /etc/wireguard/client-privatekey | wg pubkey > /etc/wireguard/client-publickey

# generate server configuration
cat > /etc/wireguard/wg0.conf <<EOF
[Interface]
Address =
PrivateKey = $(cat /etc/wireguard/server-privatekey)
ListenPort = 51820

[Peer]
PublicKey = $(cat /etc/wireguard/client-publickey)
AllowedIPs =
EOF

# generate client configuration
cat > /etc/wireguard/client.conf <<EOF
[Interface]
Address =
PrivateKey = $(cat /etc/wireguard/client-privatekey)

[Peer]
PublicKey = $(cat /etc/wireguard/server-publickey)
Endpoint =
AllowedIPs =
EOF

# start wireguard
wg-quick up wg0

# enable wireguard
systemctl enable wg-quick@wg0

# setup iptables
iptables -A INPUT -i wg0 -j ACCEPT
iptables -A FORWARD -i wg0 -j ACCEPT
iptables -A INPUT -p udp --dport 51820 -j ACCEPT
iptables -A FORWARD -i wg0 -j ACCEPT
iptables -A FORWARD -o wg0 -j ACCEPT
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE

# save iptables
iptables-save > /etc/iptables/rules.v4

# enable iptables
systemctl enable netfilter-persistent
systemctl start netfilter-persistent

# setup dnsmasq
apt-get install -y dnsmasq

# configure dnsmasq
cat > /etc/dnsmasq.conf <<EOF
interface=wg0
dhcp-range=
EOF

# restart dnsmasq
systemctl restart dnsmasq

# enable dnsmasq
systemctl enable dnsmasq

