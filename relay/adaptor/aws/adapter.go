package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/anthropic"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

var _ adaptor.Adaptor = new(Adaptor)

type Adaptor struct {
	meta      *meta.Meta
	awsClient *bedrockruntime.Client
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.meta = meta
	// Use the client pool instead of creating a new client
	pool := GetClientPool()
	a.awsClient = pool.GetClient(meta.Config.Region, meta.Config.AK, meta.Config.SK)
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return "", nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	c.Set(ctxkey.RequestModel, request.Model)
	
	// Check if it's a Llama model
	if strings.HasPrefix(request.Model, "llama-") {
		// For Llama models, store the original request since we'll convert it in the handler
		c.Set(ctxkey.ConvertedRequest, request)
		return request, nil
	} else {
		// For Claude models, convert to Anthropic format
		claudeReq := anthropic.ConvertRequest(*request)
		c.Set(ctxkey.ConvertedRequest, claudeReq)
		return claudeReq, nil
	}
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return nil, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, usage = StreamHandler(c, a.awsClient)
	} else {
		err, usage = Handler(c, a.awsClient, meta.ActualModelName)
	}
	return
}

func (a *Adaptor) GetModelList() (models []string) {
	for n := range awsModelIDMap {
		models = append(models, n)
	}
	return
}

func (a *Adaptor) GetChannelName() string {
	return "aws"
}
