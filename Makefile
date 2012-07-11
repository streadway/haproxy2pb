all: haproxy.pb.go
	cd haproxy2pb && go get -d && go build

test: all
	cat test.log | haproxy2pb/haproxy2pb

%.pb.go:	%.proto
	protoc --go_out=. $<

.PHONY: all
