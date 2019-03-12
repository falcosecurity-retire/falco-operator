IMAGE = falcosecurity/falco-operator-helm
# Use same version than helm chart
VERSION = 0.5.6

.PHONY: build bundle.yaml

build:
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
	sed -i 's|REPLACE_IMAGE|docker.io/$(IMAGE):$(VERSION)|g' deploy/operator.yaml
	cat deploy/operator.yaml >> bundle.yaml

e2e: bundle.yaml
	kubectl apply -f bundle.yaml
	kubectl apply -f deploy/crds/falco_v1alpha1_falco_cr.yaml


e2e-clean: bundle.yaml
	kubectl delete -f deploy/crds/falco_v1alpha1_falco_cr.yaml
	kubectl delete -f bundle.yaml
