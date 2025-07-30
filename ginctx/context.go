package ginctx

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"

	"github.com/fiamma-chain/fiamma-go-sdk/errors"
	"github.com/fiamma-chain/fiamma-go-sdk/log"
	"github.com/fiamma-chain/fiamma-go-sdk/utils"
)

const (
	LogKeyRequestID = "requestID"
)

// Context context
type Context struct {
	*gin.Context
	*log.Logger
}

// NewContext create a new context with gin context
func NewHttpContext(inner *gin.Context) *Context {
	return &Context{inner, log.With(log.Any(LogKeyRequestID, uuid.NewV4().String()))}
}

// LoadBody loads json data from body into object and set defaults
func (c *Context) LoadBody(obj interface{}) error {
	err := c.BindJSON(obj)
	if err != nil {
		return err
	}
	return utils.SetDefaults(obj)
}

type sucResponse struct {
	Success bool `json:"success"`
}

// PackageResponse PackageResponse
func PackageResponse(res interface{}) (int, interface{}) {
	if res == nil {
		res = &sucResponse{
			Success: true,
		}
	}
	return http.StatusOK, res
}

// PopulateFailedResponse PopulateFailedResponse
func PopulateFailedResponse(cc *Context, err error, abort bool) {
	var code string
	var status int
	switch e := err.(type) {
	case errors.Coder:
		code = e.Code()
		status = getHTTPStatus(Code(e.Code()))
	default:
		code = ErrUnknown
		status = http.StatusInternalServerError
	}

	//cc.Logger.Info("process failed.", log.Code(err))

	body := gin.H{
		"code":    code,
		"message": err.Error(),
	}
	if abort {
		cc.AbortWithStatusJSON(status, body)
	} else {
		cc.JSON(status, body)
	}
}

// HandlerFunc HandlerFunc
type HandlerFunc func(c *Context) (interface{}, error)

// Wrapper Wrapper
func Wrapper(handler HandlerFunc) func(c *gin.Context) {
	return func(c *gin.Context) {
		cc := NewHttpContext(c)
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = errors.CodeError(ErrUnknown, "failed to recover from panic")
				}
				cc.Logger.Info("handle a panic", log.Code(err), log.Error(err), log.Any("panic", string(debug.Stack())))
				PopulateFailedResponse(cc, err, false)
			}
		}()
		res, err := handler(cc)
		if err != nil {
			//cc.Logger.Info("failed to handler request", log.Any("error", err.Error()))
			PopulateFailedResponse(cc, err, false)
			return
		}
		cc.Logger.Debug("process success", log.Any("response", _toJsonString(res)))
		// unlike JSON, does not replace special html characters with their unicode entities. eg: JSON(&)->'\u0026' PureJSON(&)->'&'
		cc.PureJSON(PackageResponse(res))
	}
}

func _toJsonString(obj interface{}) string {
	data, _ := json.Marshal(obj)
	return string(data)
}
