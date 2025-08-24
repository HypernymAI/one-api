package main

import (
	"encoding/json"
	"fmt"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

func main() {
	// Test 1: Check if GoogleOpenAI is within bounds of ChannelBaseURLs
	fmt.Printf("GoogleOpenAI constant value: %d\n", channeltype.GoogleOpenAI)
	fmt.Printf("ChannelBaseURLs length: %d\n", len(channeltype.ChannelBaseURLs))
	
	// Test 2: Create a test channel with config
	channel := &model.Channel{
		Id:     1,
		Type:   channeltype.GoogleOpenAI,
		Name:   "Test GoogleOpenAI",
		Config: `{"project_id": "test-project", "region": "us-central1"}`,
	}
	
	// Test 3: Load config
	cfg, err := channel.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
	} else {
		fmt.Printf("Loaded config: %+v\n", cfg)
		fmt.Printf("ProjectID: '%s'\n", cfg.ProjectID)
		fmt.Printf("Region: '%s'\n", cfg.Region)
	}
	
	// Test 4: Check JSON marshaling/unmarshaling
	testConfig := model.ChannelConfig{
		ProjectID: "test-project",
		Region:    "us-central1",
	}
	
	jsonBytes, _ := json.Marshal(testConfig)
	fmt.Printf("\nMarshaled JSON: %s\n", string(jsonBytes))
	
	var unmarshaledConfig model.ChannelConfig
	err = json.Unmarshal(jsonBytes, &unmarshaledConfig)
	if err != nil {
		fmt.Printf("Error unmarshaling: %v\n", err)
	} else {
		fmt.Printf("Unmarshaled config: %+v\n", unmarshaledConfig)
	}
}