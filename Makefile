setup:
	go get -u github.com/Masterminds/glide
	go install github.com/Masterminds/glide

update: setup
	glide cc
	glide update --strip-vendor

get: setup
	glide install --strip-vendor

pathing:
	@if grep -lr "kastillo" . | grep -v "Makefile"; then\
		echo "ERROR: legacy import paths present in files:";\
		grep -lr "kastillo" . | grep -v "Makefile";\
		exit 1;\
	fi

test: get pathing
	go test -cover -p 1 ./...
