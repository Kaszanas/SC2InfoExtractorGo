PWD := `pwd`

DEPLOY_DIR = ./docker
TEST_COMPOSE = $(DEPLOY_DIR)/docker-dev-compose.yml

# REVIEW: Should this be ran with Docker Compose instead?
process_replays: ## Runs the container to process replays.
	docker run \
		-v "${PWD}/replays:/replays" \
		-v "${PWD}/logs:/logs" \
		-v "${PWD}/operation_files:/operation_files" \
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
	docker run --rm -v .:/app -w /app golangci/golangci-lint:latest golangci-lint run -v

###################
#### TESTING ######
###################
test_locally:
	go test ./... -v -race

compose_build_dev:
	docker-compose -f $(TEST_COMPOSE) build

compose_run_dev_it:
	docker-compose -f $(TEST_COMPOSE) run -it --rm sc2infoextractorgo-dev

compose_run_dev: compose_build_dev compose_run_dev_it

action_compose_test: ## Runs the tests in a container.
	docker-compose -f $(TEST_COMPOSE) run --rm sc2infoextractorgo-test

compose_remove: ## Stops and removes the testing containers, images, volumes.
	docker-compose -f $(TEST_COMPOSE) down --volumes --remove-orphans

compose_test: compose_build_dev action_compose_test compose_remove

.PHONY: help
help: ## Show available make targets.
	@awk '/^[^\t ]*:.*?##/{sub(/:.*?##/, ""); printf "\033[36m%-30s\033[0m %s\n", $$1, substr($$0, index($$0,$$2))}' $(MAKEFILE_LIST)
