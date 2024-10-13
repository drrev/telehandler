.DEFAULT_GOAL := build
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
SSL_DIR := ${ROOT_DIR}ssl

root-ca:
	cd ${SSL_DIR} \
	&& cfssl selfsign -config cfssl.json --profile rootca "Local Testing" csr.json | cfssljson -bare root

server-cert: root-ca
	cd ${SSL_DIR} \
	&& cfssl genkey csr_server.json | cfssljson -bare server \
	&& cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile server server.csr | cfssljson -bare server

client-cert: root-ca
	cd ${SSL_DIR} \
	&& cfssl genkey csr_client.json | cfssljson -bare client \
	&& cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile client client.csr | cfssljson -bare client \
	&& cfssl genkey csr_bubba.json | cfssljson -bare bubba \
	&& cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile bubba bubba.csr | cfssljson -bare bubba

certs: root-ca server-cert client-cert

docs:
	go run -v -tags docs . gen

build:
	go build -v -o ./telehandler .

clean:
	rm -f ssl/*.{pem,csr}
	rm -f telehandler