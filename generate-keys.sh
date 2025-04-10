openssl genrsa -out server.key 2048
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650 -subj "/CN=localhost" -addext "subjectAltName=IP:127.0.0.1"