package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// ClientPool manages a pool of AWS Bedrock clients
type ClientPool struct {
	mu      sync.RWMutex
	clients map[string]*bedrockruntime.Client
}

var (
	clientPool *ClientPool
	poolOnce   sync.Once
)

// GetClientPool returns the singleton client pool
func GetClientPool() *ClientPool {
	poolOnce.Do(func() {
		clientPool = &ClientPool{
			clients: make(map[string]*bedrockruntime.Client),
		}
	})
	return clientPool
}

// GetClient returns a cached client or creates a new one
func (p *ClientPool) GetClient(region, accessKey, secretKey string) *bedrockruntime.Client {
	key := fmt.Sprintf("%s:%s", region, accessKey)
	
	// Try to get existing client with read lock
	p.mu.RLock()
	if client, ok := p.clients[key]; ok {
		p.mu.RUnlock()
		return client
	}
	p.mu.RUnlock()
	
	// Create new client with write lock
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Double-check after acquiring write lock
	if client, ok := p.clients[key]; ok {
		return client
	}
	
	// Create new client
	client := bedrockruntime.New(bedrockruntime.Options{
		Region:      region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	})
	
	p.clients[key] = client
	return client
}

// Clear removes all clients from the pool (useful for testing or config changes)
func (p *ClientPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clients = make(map[string]*bedrockruntime.Client)
}