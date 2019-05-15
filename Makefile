IMAGE = falcosecurity/falco-operator-helm
# Use same version than helm chart
PREVIOUS_VERSION = $(shell ls -d deploy/olm-catalog/falco-operator/*/ -t | head -n1 | cut -d"/" -f4)
VERSION = 0.7.6

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
new-upstream: bundle.yaml build push operatorhub
	sed -i "s/^VERSION = .*/VERSION = $(VERSION)/" Makefile
	git add bundle.yaml
	git add Makefile
	git commit -m "New Falco helm chart release $(VERSION)"
	git tag v$(VERSION)-falco-operator-helm
	git push
	git push --tags

operatorhub:
	mkdir -p deploy/olm-catalog/falco-operator/$(VERSION)
	cp deploy/olm-catalog/falco-operator/falco-operator.template.clusterserviceversion.yaml deploy/olm-catalog/falco-operator/$(VERSION)/falco-operator.v$(VERSION).clusterserviceversion.yaml
	sed -i "s/PREVIOUS_VERSION/$(PREVIOUS_VERSION)/" deploy/olm-catalog/falco-operator/$(VERSION)/falco-operator.v$(VERSION).clusterserviceversion.yaml
	sed -i "s/VERSION/$(VERSION)/" deploy/olm-catalog/falco-operator/$(VERSION)/falco-operator.v$(VERSION).clusterserviceversion.yaml
	git add deploy/olm-catalog/falco-operator/$(VERSION)/falco-operator.v$(VERSION).clusterserviceversion.yaml

# Smoke testing targets
e2e: bundle.yaml
	kubectl apply -f bundle.yaml
	kubectl apply -f deploy/crds/falco_v1alpha1_falco_cr.yaml

e2e-clean: bundle.yaml
	kubectl delete -f deploy/crds/falco_v1alpha1_falco_cr.yaml
	kubectl delete -f bundle.yaml
