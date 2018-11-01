setup:
	go get -u github.com/Masterminds/glide
	go install github.com/Masterminds/glide

update: setup
	glide cc
	glide update --strip-vendor

get: setup
	glide install --strip-vendor

test: get
	go test -cover -p 1 ./...
