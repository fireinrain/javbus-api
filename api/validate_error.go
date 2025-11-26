package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// QueryValidationError 自定义错误响应结构
// 对应 TS: class QueryValidationError
type QueryValidationError struct {
	Error    string   `json:"error"`
	Messages []string `json:"messages"`
}

// HandleValidationError 处理验证错误并返回 JSON 响应
// 如果返回 true，表示有错误且已处理（已发送响应）；返回 false 表示无验证错误
func HandleValidationError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	var messages []string

	// 尝试将错误断言为 validator.ValidationErrors
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			// 获取字段名 (首字母小写以匹配前端 query 参数)
			field := strings.ToLower(e.Field())

			// 根据 Tag 生成友好的错误信息
			// 试图复刻 Node.js 代码中的 .withMessage(...)
			var msg string
			switch e.Tag() {
			case "required":
				msg = fmt.Sprintf("`%s` is required", field)
			case "required_with":
				msg = fmt.Sprintf("`%s` is required when `%s` is present", field, strings.ToLower(e.Param()))
			case "oneof":
				// e.Param() 会返回 "star genre director..."，我们需要格式化一下
				options := strings.Split(e.Param(), " ")
				formattedOptions := strings.Join(options, ", ")
				msg = fmt.Sprintf("`%s` must be one of [%s]", field, formattedOptions)
			case "min":
				msg = fmt.Sprintf("`%s` must be greater than or equal to %s", field, e.Param())
			case "number":
				msg = fmt.Sprintf("`%s` must be a number", field)
			default:
				msg = fmt.Sprintf("`%s` is invalid (%s)", field, e.Tag())
			}
			messages = append(messages, msg)
		}
	} else {
		// 如果不是验证错误（比如 json 解析失败），直接使用原始错误
		messages = append(messages, err.Error())
	}

	// 发送 400 响应
	c.JSON(http.StatusBadRequest, QueryValidationError{
		Error:    "query is invalid",
		Messages: messages,
	})

	return true
}
