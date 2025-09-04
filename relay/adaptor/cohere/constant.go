package cohere

var ModelList = []string{
	"command", "command-nightly",
	"command-light", "command-light-nightly",
	"command-r", "command-r-plus",
	"command-a-03-2025",
	"command-a-reasoning-08-2025",
}

func init() {
	num := len(ModelList)
	for i := 0; i < num; i++ {
		ModelList = append(ModelList, ModelList[i]+"-internet")
	}
}
