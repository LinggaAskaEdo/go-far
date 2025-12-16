package errors

import (
	"fmt"
	"net/http"
	"strings"

	"go-far/src/preference"
)

type AppError struct {
	Code       Code    `json:"code"`
	Message    string  `json:"message"`
	DebugError *string `json:"debug,omitempty"`
	sys        error
}

func init() {
	svcError = map[ServiceType]ErrorMessage{
		COMMON: ErrorMessages,
	}
}

func Compile(service ServiceType, err error, lang string, debugMode bool) (int, AppError) {
	var debugErr *string

	if debugMode {
		errStr := err.Error()
		if len(errStr) > 0 {
			debugErr = &errStr
		}
	}

	code := ErrCode(err)

	if errMessage, ok := svcError[COMMON][code]; ok {
		msg := errMessage.ID
		if lang == preference.LANG_EN {
			msg = errMessage.EN
		}

		return errMessage.StatusCode, AppError{
			Code:       code,
			Message:    msg,
			sys:        err,
			DebugError: debugErr,
		}
	}

	if errMessages, ok := svcError[service]; ok {
		if errMessage, ok := errMessages[code]; ok {
			msg := errMessage.ID
			if lang == preference.LANG_EN {
				msg = errMessage.EN
			}

			if errMessage.HasAnnotation {
				args := fmt.Sprintf("%q", err.Error())
				if start, end := strings.LastIndex(args, `{{`), strings.LastIndex(args, `}}`); start > -1 && end > -1 {
					args = strings.TrimSpace(args[start+2 : end])
					msg = fmt.Sprintf(msg, args)
				} else {
					index := strings.Index(args, `\n`)
					if index > 0 {
						args = strings.TrimSpace(args[1:index])
					}

					msg = fmt.Sprintf(msg, args)
				}
			}

			if code == CodeHTTPValidatorError {
				if err.Error() != "" {
					msg = strings.Split(err.Error(), "\n ---")[0]
				}
			}

			return errMessage.StatusCode, AppError{
				Code:       code,
				Message:    msg,
				sys:        err,
				DebugError: debugErr,
			}
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
