## Build from source

Our working solution was built by using ```go build``` command on 64 bit version of Windows 10

## Testing

The software contains defined unit tests. In order to run them please use the following command with the default provided ```.SC2Replay``` files:

```go test ./...```

If You have Your own replay files please either create directories containing different sets of replays to be tested and edit the ```TEST_INPUT_REPLAYPACK_DIR``` variable in ```./dataproc/dataproc_pipeline_test.go``` or provide them all in the ```./test_files/test_replays``` directory as input. If the ```TEST_INPUT_REPLAYPACK_DIR``` is not set, then the test is skipped. If You decide that it is suitable to skip certain directories while testing on replaypacks please include the directory names in ```TEST_BYPASS_THESE``` variable in ```./dataproc/dataproc_pipeline_test.go```.

If You wish to run the pipeline tests against a very big set of replays please keep in mind that the default Golang test timeout is set to 10 minutes. We have found that the processing of 47 tournaments from 2016 until 2021 takes about 240 minutes to complete. Example command:

```go test ./dataproc -v -run TestPipelineWrapperMultiple -timeout 240m```