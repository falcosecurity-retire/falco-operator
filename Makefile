IMAGE = falcosecurity/falco-operator
# Use same version than helm chart
VERSION = helm-based-v0.5.6

.PHONY: build bundle.yaml

build:
	operator-sdk build $(IMAGE):$(VERSION)

push:
	docker push $(IMAGE):$(VERSION)

bundle.yaml:
	cat deploy/crds/falco_v1alpha1_falco_crd.yaml > bundle.yaml
	echo '---' >> bundle.yaml
	cat deploy/service_account.yaml  >> bundle.yaml
	echo '---' >> bundle.yaml
	cat deploy/role_binding.yaml >> bundle.yaml
	echo '---' >> bundle.yaml
	sed -i 's|REPLACE_IMAGE|docker.io/$(IMAGE):$(VERSION)|g' deploy/operator.yaml
	cat deploy/operator.yaml >> bundle.yaml

e2e: bundle.yaml
	kubectl apply -f bundle.yaml
	kubectl apply -f deploy/crds/falco_v1alpha1_falco_cr.yaml


e2e-clean: bundle.yaml
	kubectl delete -f deploy/crds/falco_v1alpha1_falco_cr.yaml
	kubectl delete -f bundle.yaml
