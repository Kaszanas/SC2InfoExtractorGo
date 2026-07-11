PWD := `pwd`

DEPLOY_DIR = ./docker
TEST_COMPOSE = $(DEPLOY_DIR)/docker-dev-compose.yml
PROJECT_NAME = sc2infoextractorgo

# REVIEW: Should this be ran with Docker Compose instead?
process_replays: ## Runs the container to process replays.
	docker run \
		-v "${PWD}/replays:/replays" \
		-v "${PWD}/logs:/logs" \
		sc2infoextractorgo \
		-log_level 6


###################
#### DOCKER #######
###################
docker_build: ## Builds the "production" container.
	docker build --tag=sc2infoextractorgo -f ./docker/Dockerfile .

docker_build_dev: ## Builds the dev container.
	docker build --tag=sc2infoextractorgo:dev -f ./docker/Dockerfile.dev .

docker_run_dev: ## Runs the interactive shell in the dev container. Runs bash by default.
	docker run -it sc2infoextractorgo:dev

docker_go_lint: ## Runs the linter using the golangci-lint container.
	docker run \
			--rm \
			-v .:/app \
			-w /app \
			golangci/golangci-lint:latest \
			golangci-lint run -v --timeout 5m

###################
#### TESTING ######
###################
FIXTURE_REPLAYS_IMAGE = kaszanas/sc2replaytestdata:latest
FIXTURE_MAPS_IMAGE = kaszanas/sc2reset_maps_mods:latest

fetch_test_fixtures: ## Populates test_files/test_replays/ and dependencies/ from published fixture images.
	mkdir -p test_files/test_replays dependencies/maps dependencies/other_dependencies
	docker create --name sc2_fixture_replays $(FIXTURE_REPLAYS_IMAGE)
	docker cp sc2_fixture_replays:/sc2replaytestdata/. test_files/test_replays/
	docker rm sc2_fixture_replays
	docker create --name sc2_fixture_maps $(FIXTURE_MAPS_IMAGE)
	docker cp sc2_fixture_maps:/sc2reset_maps_mods/maps/cn_maps/. dependencies/maps/
	docker cp sc2_fixture_maps:/sc2reset_maps_mods/maps/sc2reset_maps/. dependencies/maps/
	docker cp sc2_fixture_maps:/sc2reset_maps_mods/other_dependencies/. dependencies/other_dependencies/
	docker rm sc2_fixture_maps

test_locally:
	go test ./... -v -race

compose_build_dev:
	docker compose -p $(PROJECT_NAME) -f $(TEST_COMPOSE) build

compose_run_dev_it:
	docker compose \
			-p $(PROJECT_NAME) \
			-f $(TEST_COMPOSE) \
			run -it --rm sc2infoextractorgo-dev

compose_run_dev: compose_build_dev compose_run_dev_it

action_compose_test: ## Runs the tests in a container.
	docker compose \
			-p $(PROJECT_NAME) \
			-f $(TEST_COMPOSE) \
			run --rm sc2infoextractorgo-test

compose_remove: ## Stops and removes the testing containers, images, volumes.
	docker compose \
			-p $(PROJECT_NAME) \
			-f $(TEST_COMPOSE) \
			down --volumes --remove-orphans

compose_test: compose_build_dev action_compose_test compose_remove

.PHONY: help
help: ## Show available make targets.
	@awk '/^[^\t ]*:.*?##/{sub(/:.*?##/, ""); printf "\033[36m%-30s\033[0m %s\n", $$1, substr($$0, index($$0,$$2))}' $(MAKEFILE_LIST)
