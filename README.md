## Objects that will have to be anonymized

- `replayFile.InitData.LobbyState.Slots` - This needs to be checked

- `replayFile.InitData.GameDescription.UserInitDatas`

- `replayFile.Details.players`


## Doc references to the variables that need to be anonymized:

1. Player - Holds name and toon id (it needs to be veryfied if its populating through the structure)
https://godoc.org/github.com/icza/s2prot/rep#Player

## Working with MPQ Files:
Possible MPQ packages that might help with anonymization process without prior replay operations:

1. https://lib.rs/crates/mpqtool - Rust based command line tool for working with MPQ

2. http://www.zezula.net/en/mpq/download.html - MPQ Editor (Command line usage is not clear for this one)

### Blizzard anonymized vs unanonymized MPQ

It was observed that Blizzard while anonymizing replays deleted some archives from MPQ.

Anonymized replay contains:
- replay.attributes.events
- replay.details.backup
- replay.game.events
- replay.gamemetadata.json
- replay.initData.backup
- replay.load.info

Not anonymized replay contains:
- replay.attributes.events
- replay.details
- replay.details.backup
- replay.game.events
- replay.gamemetadata.json
- replay.initData
- replay.initData.backup
- replay.load.info
- replay.message.events
- replay.resumable.events
- replay.server.battlelobby
- replay.smartcam.events
- replay.sync.events
- replay.sync.history
- replay.tracker.events

So the difference is:
- replay.details
- replay.initData
- replay.message.events
- replay.resumable.events
- replay.server.battlelobby
- replay.smartcam.events
- replay.sync.events
- replay.sync.history
- replay.tracker.events
