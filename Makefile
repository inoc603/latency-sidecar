build:
	GOOS=linux GOARCH=amd64 go build .

image: build
	docker build -f Dockerfile -t latency-sidecar .

k8s/run: image
	kubectl create -f pods.yml

k8s/stop:
	kubectl delete pods latency-test

k8s/test:
	docker build -f Dockerfile.poc -t latency-test .
	@kubectl run -it --rm \
		--restart=Never \
		--image=latency-test \
		--image-pull-policy=Never \
		client -- \
		sh -c 'sleep 2 && ./test.sh $(shell kubectl get pods latency-test -o go-template='{{.status.podIP}}') 8080 20ms'

run: build
	docker-compose build && docker-compose up
