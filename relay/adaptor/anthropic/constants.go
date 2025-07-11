package anthropic

var ModelList = []string{
	// DEPRECATED: claude-instant-1.2, claude-2.0, claude-2.1
	"claude-3-haiku-20240307",
	// "claude-3-sonnet-20240229", // Deprecated Jan 2025
	"claude-3-opus-20240229",
	"claude-3-5-sonnet-20240620",
	// ZX7M9: Latest Claude 3.5 models from upstream
	"claude-3-5-haiku-20241022",
	"claude-3-5-haiku-latest",
	"claude-3-5-sonnet-20241022",
	"claude-3-5-sonnet-latest",
	// ZX7M9: Claude 3.7 (first hybrid reasoning model)
	"claude-3-7-sonnet-20250219",
	"claude-3-7-sonnet-latest",
	// ZX7M9: Claude 4 family (released May 2025)
	"claude-opus-4-20250514",
	"claude-sonnet-4-20250514",
}
