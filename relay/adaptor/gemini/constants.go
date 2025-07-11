package gemini

// https://ai.google.dev/models/gemini

var ModelList = []string{
	// "gemini-pro", // DEPRECATED - use gemini-1.5-pro instead
	// "gemini-1.0-pro", // DEPRECATED - use gemini-1.5-pro instead
	"gemini-1.5-pro",
	"gemini-pro-vision", "gemini-1.0-pro-vision-001",
	// ZX7M9: Missing models from upstream
	"gemini-1.5-flash", "gemini-1.5-flash-8b", "gemini-1.5-pro-experimental",
	"gemini-2.0-flash",
	"gemini-2.0-flash-exp",
	"gemini-2.0-flash-lite-preview-02-05",
	"gemini-2.0-flash-thinking-exp-01-21",
	"gemini-2.0-pro-exp-02-05",
	"text-embedding-004", "aqa",
	// ZX7M9: Gemini 2.5 models (GA June 2025)
	"gemini-2.5-flash",
	"gemini-2.5-pro",
}
