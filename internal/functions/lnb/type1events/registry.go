package type1events

var LNBType1Events = map[string]Action{
	BuildKey("50", "000000"): HandleEvent50000000,
	BuildKey("04", "000000"): HandleEvent04000000,
}
