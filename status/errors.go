package status

import "fmt"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) WithMessage(msg string) *Error {
	message := e.Message
	if message != "" {
		message = message + ": " + msg
	} else {
		message = msg
	}
	return &Error{
		Code:    e.Code,
		Message: message,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

var (
	ErrorCodeApiFileExist  = &Error{Code: 10001, Message: "该目录或文件已存在"}
	ErrorCodeApiBuildError = &Error{Code: 10002, Message: "项目重新编译失败"}
)
