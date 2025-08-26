package openai

var ModelList = []string{
	"gpt-3.5-turbo", "gpt-3.5-turbo-0301", "gpt-3.5-turbo-0613", "gpt-3.5-turbo-1106", "gpt-3.5-turbo-0125",
	"gpt-3.5-turbo-16k", "gpt-3.5-turbo-16k-0613",
	"gpt-3.5-turbo-instruct",
	"gpt-4", "gpt-4-0314", "gpt-4-0613", "gpt-4-1106-preview", "gpt-4-0125-preview",
	"gpt-4-32k", "gpt-4-32k-0314", "gpt-4-32k-0613",
	"gpt-4-turbo-preview", "gpt-4-turbo", "gpt-4-turbo-2024-04-09",
	"gpt-4o", "gpt-4o-2024-05-13",
	"gpt-4o-mini", "gpt-4o-mini-2024-07-18",
	"gpt-4-vision-preview", "gpt-4.1-nano", "gpt-4.1-mini", "gpt-4.1", "gpt-4.5-preview",
	"text-embedding-ada-002", "text-embedding-3-small", "text-embedding-3-large",
	"text-curie-001", "text-babbage-001", "text-ada-001", "text-davinci-002", "text-davinci-003",
	"text-moderation-latest", "text-moderation-stable",
	"text-davinci-edit-001",
	"davinci-002", "babbage-002",
	"dall-e-2", "dall-e-3",
	"whisper-1",
	"tts-1", "tts-1-1106", "tts-1-hd", "tts-1-hd-1106",
	// ZX7M9: O3/O4 models (April 2025 release)
	"o3", "o3-2025-04-16",
	"o3-mini", "o3-mini-2025-01-31",
	"o4-mini", "o4-mini-2025-04-16",
	// GPT-5 series models
	"gpt-5-chat", "gpt-5-mini", "gpt-5-nano",
}
