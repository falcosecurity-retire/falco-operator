IMAGE = falcosecurity/falco-operator
# Use same version than helm chart
VERSION = helm-based-v0.5.6

.PHONY: build

build:
	operator-sdk build $(IMAGE):$(VERSION)

push:
	docker push $(IMAGE):$(VERSION)

e2e:
	kubectl apply -f deploy/crds/falco_v1alpha1_falco_crd.yaml
	sed -i 's|REPLACE_IMAGE|docker.io/$(IMAGE):$(VERSION)|g' deploy/operator.yaml
	kubectl apply -f deploy/service_account.yaml
	kubectl apply -f deploy/role_binding.yaml
	kubectl apply -f deploy/operator.yaml
	kubectl apply -f deploy/crds/falco_v1alpha1_falco_cr.yaml


e2e-clean:
	kubectl delete -f deploy/crds/falco_v1alpha1_falco_cr.yaml
	kubectl delete -f deploy/operator.yaml
	kubectl delete -f deploy/role_binding.yaml
	kubectl delete -f deploy/service_account.yaml
	kubectl delete -f deploy/crds/falco_v1alpha1_falco_crd.yaml
