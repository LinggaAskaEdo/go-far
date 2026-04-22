package errors

import (
	"fmt"
	"net/http"
	"strings"

	"go-far/src/preference"
)

type AppError struct {
	sys        error
	DebugError *string `json:"debug,omitempty"`
	Message    string  `json:"message"`
	Code       Code    `json:"code"`
}

func init() {
	svcError = map[ServiceType]ErrorMessage{
		COMMON: ErrorMessages,
	}
}

func Compile(service ServiceType, err error, lang string, debugMode bool) (int, AppError) {
	debugErr := getDebugError(err, debugMode)
	code := ErrCode(err)

	if errMessage, ok := svcError[COMMON][code]; ok {
		return buildAppError(errMessage, code, err, lang, debugErr)
	}

	if errMessages, ok := svcError[service]; ok {
		if errMessage, ok := errMessages[code]; ok {
			return buildAppErrorWithAnnotation(errMessage, code, err, lang, debugErr)
		}

		return http.StatusInternalServerError, AppError{
			Code:       code,
			Message:    "error message not defined!",
			sys:        err,
			DebugError: debugErr,
		}
	}

	return http.StatusInternalServerError, AppError{
		Code:       code,
		Message:    "service error not defined!",
		sys:        err,
		DebugError: debugErr,
	}
}

func getDebugError(err error, debugMode bool) *string {
	if !debugMode {
		return nil
	}
	errStr := err.Error()

	if errStr == "" {
		return nil
	}

	return &errStr
}

func getMessage(msg Message, lang string) string {
	result := msg.ID
	if lang == preference.LANG_EN {
		result = msg.EN
	}

	return result
}

func buildAppError(msg Message, code Code, err error, lang string, debugErr *string) (int, AppError) {
	return msg.StatusCode, AppError{
		Code:       code,
		Message:    getMessage(msg, lang),
		sys:        err,
		DebugError: debugErr,
	}
}

func formatAnnotation(msg, errStr string) string {
	args := fmt.Sprintf("%q", errStr)
	start, end := strings.LastIndex(args, `{{`), strings.LastIndex(args, `}}`)
	if start > -1 && end > -1 {
		args = strings.TrimSpace(args[start+2 : end])
		return fmt.Sprintf(msg, args)
	}

	index := strings.Index(args, `\n`)
	if index > 0 {
		args = strings.TrimSpace(args[1:index])
	}

	return fmt.Sprintf(msg, args)
}

func buildAppErrorWithAnnotation(msg Message, code Code, err error, lang string, debugErr *string) (int, AppError) {
	result := getMessage(msg, lang)
	errStr := err.Error()

	if msg.HasAnnotation {
		result = formatAnnotation(result, errStr)
	}

	if code == CodeHTTPValidatorError && errStr != "" {
		result = strings.Split(errStr, "\n ---")[0]
	}

	return msg.StatusCode, AppError{
		Code:       code,
		Message:    result,
		sys:        err,
		DebugError: debugErr,
	}
}
