IMAGE = falcosecurity/falco-operator
VERSION = 0.15.0


.PHONY: testrun
testrun:
	go build . && ./falco-operator supervise --set "falco.yaml=rules_file={{.File.Path}}" --watch foo --watch-interval 1s --restart-grace-period 5s --stop-grace-period 500ms -- bash -c 'echo started myapp; cat foo; sleep 100'

.PHONY: testrun2
testrun2:
	go build . && ./falco-operator supervise --set "falco.yaml=rules_file.-1={{.Path}}+{{.Content}}" --watch watched --watch-interval 1s --restart-grace-period 5s --stop-grace-period 500ms -- bash -c 'echo started myapp; echo dumping falco.yaml...; cat falco.yaml; sleep 100'

.PHONY: server-defaultns
server-defaultns:
	OPERATOR_NAME=falco-operator operator-sdk up local

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
	docker build -t ${IMAGE}:${VERSION} .

.PHONY: clean
clean:
	rm -rf build

.PHONY: push
push: 
	docker push ${IMAGE}:${VERSION}

.PHONY: new-upstream
new-upstream: dockerimage push 
	sed -i "s/^VERSION = .*/VERSION = ${VERSION}/" Makefile
	git add Makefile
	git commit -m "New Falco operator release ${VERSION}"
	git tag -f v${VERSION}
	GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no" git push origin HEAD:master
	GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no" git push --tags -f
