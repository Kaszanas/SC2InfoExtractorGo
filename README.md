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
