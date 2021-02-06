package main

type CleanedReplay struct {
	header         CleanedHeader
	initData       CleanedInitData
	details        CleanedDetails
	metadata       CleanedMetadata
	gameEvtsErr    bool
	messageEvtsErr bool
	trackerEvtsErr bool
}
