package datastruct

var engSixteenBitLE = "16-Bit LE"
// TODO: Check this file:
var engAbiogenesis = "Abiogenesis LE"
var engAbyssalReef = "Abyssal Reef LE"

var MapNames = map[string]map[string]string{
	engSixteenBitLE: {
		// German, Italian, English
		"16-Bit LE": engSixteenBitLE,
		// Spanish
		"16 bits EE": engSixteenBitLE,
		"16 bits EJ": engSixteenBitLE,
		// French
		"16 bits EC": engSixteenBitLE,
		// Polish
		"16 bitów ER": engSixteenBitLE,
		// Portuguese
		"16 Bits LE": engSixteenBitLE,
		// Russian
		"16 бит РВ":  engSixteenBitLE,
		// Asian Regions
		"16비트 - 래더":  engSixteenBitLE,
		"16位-天梯版":    engSixteenBitLE,
		"16位元 - 天梯版": engSixteenBitLE,
	},
	// TODO: Check if Abiogenesis is having any other names. It seems that the original file does not contain locale.
	engAbiogenesis: {
		engAbiogenesis: engAbiogenesis,
	}
	engAbyssalReef: {
		engAbyssalReef: engAbyssalReef,
		// German
		"Tiefseeriff LE": engAbyssalReef,
		// Spanish
		"Arrecife abisal EE": engAbyssalReef,
		"Arrecife Abisal EJ": engAbyssalReef,
		// French
		"Récif abyssal EC": engAbyssalReef,
		// Italian
		"Barriera sommersa LE": engAbyssalReef,
		// Polish
		"Rafa otchłani ER": engAbyssalReef,
		// Portuguese
		"Recife Abissal LE": engAbyssalReef,
		// Russian
		"Глубоководный риф РВ": engAbyssalReef,
		// Asian Regions
		"어비설 리프 - 래더": engAbyssalReef,
		"深海暗礁 - 天梯版": engAbyssalReef,
		"深海礁岩 - 天梯版": engAbyssalReef,
	}
}
