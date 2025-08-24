package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/middleware"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
)

func main() {
	// Create a test channel
	channel := &model.Channel{
		Id:      1,
		Type:    channeltype.GoogleOpenAI,
		Name:    "Test GoogleOpenAI",
		Key:     "test-key",
		BaseURL: nil, // Let it use default
		Config:  `{"project_id": "test-project", "region": "us-central1"}`,
		Models:  "meta/llama-3.1-8b-instruct-maas",
	}

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Method: "POST",
		URL:    &url.URL{Path: "/v1/chat/completions"},
		Header: make(http.Header),
	}
	c.Request.Header.Set("Authorization", "Bearer test-token")

	// Step 1: Load channel config
	fmt.Println("=== Step 1: Loading channel config ===")
	cfg, err := channel.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}
	fmt.Printf("Loaded config: %+v\n", cfg)
	fmt.Printf("ProjectID: '%s', Region: '%s'\n", cfg.ProjectID, cfg.Region)

	// Step 2: Setup context (simulating distributor)
	fmt.Println("\n=== Step 2: Setting up context ===")
	c.Set(ctxkey.Channel, channel.Type)
	c.Set(ctxkey.ChannelId, channel.Id)
	c.Set(ctxkey.ChannelName, channel.Name)
	c.Set(ctxkey.BaseURL, channel.GetBaseURL())
	c.Set(ctxkey.Config, cfg)
	c.Set(ctxkey.ModelMapping, channel.GetModelMapping())
	c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))

	// Alternative: Use SetupContextForSelectedChannel
	fmt.Println("\n=== Step 2b: Using SetupContextForSelectedChannel ===")
	middleware.SetupContextForSelectedChannel(c, channel, "meta/llama-3.1-8b-instruct-maas")

	// Step 3: Get metadata
	fmt.Println("\n=== Step 3: Getting metadata ===")
	metaData := meta.GetByContext(c)
	fmt.Printf("Meta struct: %+v\n", metaData)
	fmt.Printf("Meta.Config: %+v\n", metaData.Config)
	fmt.Printf("Meta.Config.ProjectID: '%s'\n", metaData.Config.ProjectID)
	fmt.Printf("Meta.Config.Region: '%s'\n", metaData.Config.Region)
	fmt.Printf("Meta.BaseURL: '%s'\n", metaData.BaseURL)
	fmt.Printf("Meta.ChannelType: %d\n", metaData.ChannelType)

	// Step 4: Check what's in the context
	fmt.Println("\n=== Step 4: Checking context values ===")
	if ctxCfg, exists := c.Get(ctxkey.Config); exists {
		fmt.Printf("Config from context exists: %+v\n", ctxCfg)
		if typedCfg, ok := ctxCfg.(model.ChannelConfig); ok {
			fmt.Printf("Typed config: %+v\n", typedCfg)
			fmt.Printf("ProjectID: '%s', Region: '%s'\n", typedCfg.ProjectID, typedCfg.Region)
		} else {
			fmt.Printf("Config type assertion failed. Actual type: %T\n", ctxCfg)
		}
	} else {
		fmt.Println("Config not found in context!")
	}

	// Step 5: Test JSON marshaling of the config directly
	fmt.Println("\n=== Step 5: Testing JSON marshaling ===")
	jsonBytes, _ := json.Marshal(metaData.Config)
	fmt.Printf("Marshaled meta.Config: %s\n", string(jsonBytes))
}