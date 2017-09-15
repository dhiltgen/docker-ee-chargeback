UCP_VERSION=2.2
ORG=dhiltgen/

build:
	docker build -t $(ORG)chargeback:$(UCP_VERSION) .

push:
	docker push $(ORG)chargeback:$(UCP_VERSION)
