build:
	GOOS=linux GOARCH=amd64 go build .

image: build
	docker build -f Dockerfile -t latency-sidecar .

pod=pods.yml
k8s/run: image k8s/stop
	kubectl create -f $(pod)
	kubectl wait --for=condition=ready pod/latency-test

k8s/stop:
	-kubectl delete pods latency-test

skip_build=0

k8s/logs:
	kubectl logs -f latency-test -c agent

k8s/test:
ifneq ($(skip_build), 1)
	docker build -f Dockerfile.poc -t latency-test .
endif
	@kubectl run -it --rm \
		--restart=Never \
		--image=latency-test \
		--image-pull-policy=Never \
		client -- \
		sh -c 'sleep 2 && ./test.sh $(shell kubectl get pods latency-test -o go-template='{{.status.podIP}}') 8080 20ms'

run: build
	docker-compose build && docker-compose up

