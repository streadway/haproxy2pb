all: haproxy.pb.go scan.go haproxy2pb/main.go
	go install
	cd haproxy2pb && go install

test: haproxy.pb.go
	go install
	cd haproxy2pb && go install
	cd ..
	cat test.log | haproxy2pb

%.pb.go:	%.proto
	protoc --go_out=. $<

.PHONY: all
