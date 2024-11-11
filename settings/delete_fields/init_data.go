package settings

// GameDescriptionFields is a slice of fields that
// are going to be deleted from rep.Rep.InitData.GameDescription
var GameDescriptionFields = []string{}

// GameDescriptionGameOptionsFields is a slice of fields
// that are going to be deleted from rep.Rep.InitData.GameDescription.GameOptions.Struct
var GameDescriptionGameOptionsFields = []string{
	"advancedSharedControl",
	"buildCoachEnabled",
	"clientDebugFlags",
	"fog",
	"heroDuplicatesAllowed",
	"lockTeams",
	"practice",
	"randomRaces",
	"teamsTogether",
	"userDifficulty",
}
