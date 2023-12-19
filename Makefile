PWD := `pwd`

all:
	docker run \
		-v "${PWD}/DEMOS:/DEMOS" \
		-v "${PWD}/logs:/logs" \
		-v "${PWD}/operation_files:/operation_files" \
		sc2-info-extractor \
		./SC2InfoExtractorGo -log_level 6

build:
	docker build --tag=sc2-info-extractor -f ./Dockerfile .
