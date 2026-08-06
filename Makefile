.PHONY: inference-test inference-race inference-run inference-render inference-image \
	inference-compose-config inference-compose-up inference-compose-down \
	inference-helm-lint inference-helm-template inference-robot-dryrun inference-robot-smoke \
	inference-istio-render inference-observability-render \
	inference-observability-compose-config inference-observability-up inference-observability-down \
	inference-promtool inference-dashboard-validate

inference-test:
	go test -buildvcs=false ./internal/inference/... ./cmd/inference-lab ./cmd/exercise -count=1
	go vet -buildvcs=false ./internal/inference/... ./cmd/inference-lab ./cmd/exercise

inference-race:
	go test -buildvcs=false -race ./internal/inference/... -count=1

inference-run:
	go run -buildvcs=false ./cmd/inference-lab -mode=all

inference-render:
	kubectl kustomize deploy/inference-lab/base
	kubectl kustomize deploy/inference-lab/overlays/learning

inference-istio-render:
	kubectl kustomize deploy/inference-lab/istio

inference-observability-render:
	kubectl kustomize deploy/inference-lab/observability

inference-image:
	docker build -f deploy/inference-lab/Dockerfile -t inference-lab:dev .

inference-compose-config:
	docker compose -f deploy/inference-lab/compose.yaml config --quiet

inference-compose-up:
	docker compose -f deploy/inference-lab/compose.yaml up --build

inference-compose-down:
	docker compose -f deploy/inference-lab/compose.yaml down

inference-observability-compose-config:
	docker compose -f deploy/inference-lab/compose.yaml \
		-f deploy/inference-lab/observability/compose.yaml config --quiet

inference-observability-up:
	docker compose -f deploy/inference-lab/compose.yaml \
		-f deploy/inference-lab/observability/compose.yaml up --build

inference-observability-down:
	docker compose -f deploy/inference-lab/compose.yaml \
		-f deploy/inference-lab/observability/compose.yaml down

inference-promtool:
	promtool check rules deploy/inference-lab/observability/prometheus/rules.yml

inference-dashboard-validate:
	python -c "import json; json.load(open('deploy/inference-lab/observability/grafana/dashboards/inference-router.json', encoding='utf-8'))"

inference-helm-lint:
	helm lint deploy/inference-lab/chart -f deploy/inference-lab/chart/values-learning.yaml

inference-helm-template:
	helm template learning deploy/inference-lab/chart -f deploy/inference-lab/chart/values-learning.yaml

inference-robot-dryrun:
	robot --dryrun --output NONE --log NONE --report NONE test/acceptance/inference

inference-robot-smoke:
	robot --outputdir .robot-results/inference test/acceptance/inference/00_smoke.robot
