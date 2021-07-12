#!/bin/sh
#
rm *.pem

# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=FR/ST=Ile de France/L=Paris/O=Home/CN=*.scanner.gg/emailAddress=server@scanner.gg"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

# Generate server certificat
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=FR/ST=Ile de France/L=Paris/O=Home/CN=server.scanner.gg/emailAddress=server@scanner.gg"

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf
echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text

# 4. Generate client's private key and certificate signing request (CSR)
# workers
openssl req -newkey rsa:4096 -nodes -keyout client-worker-key.pem -out client-worker-req.pem -subj "/C=FR/ST=Ile de France/L=Paris/O=Home/CN=worker.scanner.gg/emailAddress=worker@scanner.gg"
# cyrinux
openssl req -newkey rsa:4096 -nodes -keyout client-cyrinux-key.pem -out client-cyrinux-req.pem -subj "/C=FR/ST=Ile de France/L=Paris/O=Home/CN=cyrinux.scanner.gg/emailAddress=cyrinux@scanner.gg"

# 5. Use CA's private key to sign client's CSR and get back the signed certificate
# workers
openssl x509 -req -in client-worker-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-worker-cert.pem
# cyrinux
openssl x509 -req -in client-cyrinux-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cyrinux-cert.pem

echo "Client's signed certificate"
openssl x509 -in client-worker-cert.pem -noout -text
openssl x509 -in client-cyrinux-cert.pem -noout -text

chmod 644 *.pem
