build:
	@sh -c "'$(CURDIR)/build.sh'"

push:
	@sh -c "'$(CURDIR)/push.sh'"

build_testing:
	docker build --tag pelldvs/testing -f ./Dockerfile.testing .

.PHONY: build push build_testing
