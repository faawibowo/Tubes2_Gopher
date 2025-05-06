package app

import (
	"github.com/faawibowo/Tubes2_Gopher/pkg/errcode"
	"github.com/faawibowo/Tubes2_Gopher/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Validation(c *gin.Context, param any, response *Response) error {
	if err := c.ShouldBind(param); err != nil {
		logger.WithTrace(c).Errorf("params errs: %v", err)
		response.ToErrorResponse(errcode.InvalidParams.WithDetails(err.Error()))
		return err
	}
	return nil
}
