# Contributing

Contributions are welcome, and they are greatly appreciated! Every little bit
helps, and credit will always be given.

## Types of Contributions

### Report Bugs

If you are reporting a bug, please include all relevant information that may point to the possible solution of the issue. This may include:

* Your operating system name and version.
* Any details about your local setup that might be helpful in troubleshooting.
* Detailed steps to reproduce the bug.

### Fix Bugs

Look through the GitHub issues for bugs. Anything there is open to whoever wants to fix it.

### Implement Features

Look through the GitHub issues for features. Anything there is open to whoever wants to implement it to improve our solution.

### Write Documentation

You can never have enough documentation! Please feel free to contribute to any
part of the documentation, such as the official docs, docstrings, or even
on the web in blog posts, articles, and such.

### Submit Feedback

If you are proposing a feature:

* Explain in detail how it would work.
* Keep the scope as narrow as possible, to make it easier to implement.
* Remember that this is a volunteer-driven project, and that contributions
  are welcome :)

## Dockerized Development

In case you want to contribute to the project, you should set up your development environment. If you wish to use Docker, you will have to clone the repository and set up the respective Docker images.

### Build Docker Images

Setting up the development environment can be a tiresome process. Due to this we have prepared a dockerfile that can be used as a devcontainer. You can find it under `docker/Dockerfile.dev`. This image should have all of the base dependencies that are required for development.

To build the image, run the following command:

```sh
make docker_build_dev
```

or if you want to build it with docker compose:

```sh
make compose_build_dev
```

> [!NOTE]
> To fully test the tool you should build both the development and the production images.
> The development image is used for testing and debugging the tool, while the production image is used for running the tool in a production environment.

To build the production image, run the following command:

```sh
make docker_build
```

## Build from source

Our working solution was built by using ```go build``` command on 64 bit version of Windows 10

## Testing

Our tool contains some defined unit tests. In order to run them please use the following command:

```go test ./...```

The default testing files will be downloaded if not available and the tests will run on the provided default testing sample.

### Testing With Custom Files

If You have Your own replay files please either create directories containing different sets of replays to be tested and edit the ```TEST_INPUT_REPLAYPACK_DIR``` variable in ```./dataproc/dataproc_pipeline_test.go``` or provide them all in the ```./test_files/test_replays``` directory as input. If the ```TEST_INPUT_REPLAYPACK_DIR``` is not set, then the test is skipped. If You decide that it is suitable to skip certain directories while testing on replaypacks please include the directory names in ```TEST_BYPASS_THESE``` variable in ```./dataproc/dataproc_pipeline_test.go```.

### Testing on Big Datasets

If You wish to run the pipeline tests against a very big set of replays please keep in mind that the default Golang test timeout is set to 10 minutes. We have found that the processing of 47 tournaments from 2016 until 2021 takes about 240 minutes to complete. Example command:

```go test ./dataproc -v -run TestPipelineWrapperMultiple -timeout 240m```