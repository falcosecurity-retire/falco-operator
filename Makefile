IMAGE = falcosecurity/falco-operator-helm
# Use same version than helm chart
VERSION = 0.7.3

.PHONY: build bundle.yaml

build:
	helm repo update
	helm fetch stable/falco --version $(VERSION) --untar --untardir helm-charts/
	operator-sdk build $(IMAGE):$(VERSION)
	rm -fr helm-charts/falco

push:
	docker push $(IMAGE):$(VERSION)

bundle.yaml:
	cat deploy/crds/falco_v1alpha1_falco_crd.yaml > bundle.yaml
	echo '---' >> bundle.yaml
	cat deploy/service_account.yaml  >> bundle.yaml
	echo '---' >> bundle.yaml
	cat deploy/role_binding.yaml >> bundle.yaml
	echo '---' >> bundle.yaml
	cat deploy/operator.yaml >> bundle.yaml
	sed -i 's|REPLACE_IMAGE|docker.io/$(IMAGE):$(VERSION)|g' bundle.yaml

# Synchronize operator with the same Helm Chart version
new-upstream: build push bundle.yaml
	sed -i "s/^VERSION = .*/VERSION = $(VERSION)/" Makefile
	git add bundle.yaml
	git add Makefile
	git commit -m "New Falco helm chart release $(VERSION)"
	git tag falco-operator-helm-v$(VERSION)
	git push --all

# Smoke testing targets
e2e: bundle.yaml
	kubectl apply -f bundle.yaml
	kubectl apply -f deploy/crds/falco_v1alpha1_falco_cr.yaml

e2e-clean: bundle.yaml
	kubectl delete -f deploy/crds/falco_v1alpha1_falco_cr.yaml
	kubectl delete -f bundle.yaml
