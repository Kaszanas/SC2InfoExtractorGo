PWD := `pwd`

process_replays: ## Runs the container to process replays.
	docker run \
		-v "${PWD}/replays:/replays" \
		-v "${PWD}/logs:/logs" \
		-v "${PWD}/operation_files:/operation_files" \
		sc2infoextractorgo \
		-log_level 6

build: ## Builds the "production" container.
	docker build --tag=sc2infoextractorgo -f ./docker/Dockerfile .

build_dev: ## Builds the dev container.
	docker build --tag=sc2infoextractorgo:dev -f ./docker/Dockerfile.dev .

run_dev: ## Runs the interactive shell in the dev container.
	docker run -it sc2infoextractorgo:dev

.PHONY: help
help: ## Show available make targets.
	@awk '/^[^\t ]*:.*?##/{sub(/:.*?##/, ""); printf "\033[36m%-30s\033[0m %s\n", $$1, substr($$0, index($$0,$$2))}' $(MAKEFILE_LIST)
