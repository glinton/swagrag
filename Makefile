.PHONY: build
build:
	go build .

.PHONY: build-docker
build-docker: 
	docker build -t docker.io/glinton/swagrag -f scripts/dockerfile .

.PHONY: publish-docker
publish-docker: build-docker
	docker push docker.io/glinton/swagrag
