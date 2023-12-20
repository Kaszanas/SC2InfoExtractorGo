PWD := `pwd`

all:
	docker run \
		-v "${PWD}/DEMOS:/DEMOS" \
		-v "${PWD}/logs:/logs" \
		-v "${PWD}/operation_files:/operation_files" \
		sc2-info-extractor \
		./SC2InfoExtractorGo -log_level 6

build:
	docker build --tag=sc2infoextractorgo -f ./docker/Dockerfile .

build_dev:
	docker build --tag=sc2infoextractorgo:dev -f ./docker/Dockerfile.dev .

run_dev:
	docker run -it sc2infoextractorgo:dev