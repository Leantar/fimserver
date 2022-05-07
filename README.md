# fimserver

```
Generate Certificates:

openssl ecparam -out ca.key -name secp384r1 -genkey
openssl req -new -sha384 -key ca.key -out ca.csr
openssl x509 -req -sha384 -days 365 -in ca.csr -signkey ca.key -out ca.pem

openssl ecparam -out server.key -name secp384r1 -genkey
openssl req -new -sha384 -key server.key -out server.csr -config server-cert.cnf
openssl x509 -req -sha384 -days 365 -in server.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out server.pem -extensions req_ext -extfile server-cert.cnf
openssl x509 -text -noout -in server.pem | grep -A 1 "Subject Alternative Name"

openssl ecparam -out admin_client.key -name secp384r1 -genkey
openssl req -new -sha384 -key admin_client.key -out admin_client.csr
openssl x509 -req -sha384 -days 365 -in admin_client.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out admin_client.pem

openssl ecparam -out agent_client.key -name secp384r1 -genkey
openssl req -new -sha384 -key agent_client.key -out agent_client.csr
openssl x509 -req -sha384 -days 365 -in agent_client.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out agent_client.pem
```