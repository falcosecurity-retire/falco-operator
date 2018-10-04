.PHONY: testrun
testrun:
	go build . && ./falco-operator supervise --set "falco.yaml=rules_file={{.File.Path}}" --watch foo --watch-interval 1s --restart-grace-period 5s --stop-grace-period 500ms -- bash -c 'echo started myapp; cat foo; sleep 100'

.PHONY: testrun2
testrun2:
	go build . && ./falco-operator supervise --set "falco.yaml=rules_file.-1={{.Path}}+{{.Content}}" --watch watched --watch-interval 1s --restart-grace-period 5s --stop-grace-period 500ms -- bash -c 'echo started myapp; echo dumping falco.yaml...; cat falco.yaml; sleep 100'

.PHONY: server-defaultns
server-defaultns:
	OPERATOR_NAME=falco-erator operator-sdk up local

# WATCH_NAMESPACE="" does not seem to work
.PHONY: server-allns
server-allns:
	OPERATOR_NAME=falco-operator operator-sdk up local --kubeconfig ~/.kube/config --namespace ""

.PHONY: build-linux
build-linux:
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o build/falco-operator-amd64 .

.PHONY: dockerimage
dockerimage: build-linux
	docker build -t mumoshu/falco-operator:v0.12.1 .
