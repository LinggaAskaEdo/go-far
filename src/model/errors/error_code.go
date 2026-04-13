package errors

import "github.com/palantir/stacktrace"

var (
	svcError map[ServiceType]ErrorMessage

	ErrCode      = stacktrace.GetCode
	New          = stacktrace.NewError
	NewWithCode  = stacktrace.NewErrorWithCode
	RootCause    = stacktrace.RootCause
	Wrap         = stacktrace.Propagate
	WrapWithCode = stacktrace.PropagateWithCode
)

type (
	ServiceType  int
	Code         = stacktrace.ErrorCode
	ErrorMessage map[Code]Message

	Message struct {
		StatusCode    int    `json:"status_code"`
		EN            string `json:"en"`
		ID            string `json:"id"`
		HasAnnotation bool
	}
)

const (
	COMMON ServiceType = 1
)

const (
	// Code HTTP Handler
	CodeHTTPBadRequest = Code(iota + 100)
	CodeHTTPNotFound
	CodeHTTPUnauthorized
	CodeHTTPInternalServerError
	CodeHTTPUnmarshal
	CodeHTTPMarshal
	CodeHTTPConflict
	CodeHTTPForbidden
	CodeHTTPUnprocessableEntity
	CodeHTTPTooManyRequest
	CodeHTTPValidatorError
	CodeHTTPServiceUnavailable
	CodeHTTPParamDecode
	CodeHTTPErrorOnReadBody
)

const (
	// Error on SQL
	CodeSQLBuilder = Code(iota + 200)
	CodeSQLRead
	CodeSQLRowScan
	CodeSQLCreate
	CodeSQLUpdate
	CodeSQLDelete
	CodeSQLUnlink
	CodeSQLTxBegin
	CodeSQLTxCommit
	CodeSQLPrepareStmt
	CodeSQLRecordMustExist
	CodeSQLCannotRetrieveLastInsertID
	CodeSQLCannotRetrieveAffectedRows
	CodeSQLUniqueConstraint
	CodeSQLRecordDoesNotMatch
	CodeSQLRecordIsExpired
	CodeSQLRecordDoesNotExist
	CodeSQLForeignKeyMissing
	CodeSQLTxRollback
	CodeRequestIDIsNotMatch
	CodeSQLConflict
	CodeSQLEmptyRow
	CodeSQLTableNotExist
	CodeSQLQueryBuild
)

const (
	// Error on Token
	CodeTokenStillValid = Code(iota + 300)
	CodeTokenRefreshStillValid
)

const (
	// Error On Cache
	CodeCacheMarshal = Code(iota + 400)
	CodeCacheUnmarshal
	CodeCacheGetSimpleKey
	CodeCacheSetSimpleKey
	CodeCacheDeleteSimpleKey
	CodeCacheGetHashKey
	CodeCacheSetHashKey
	CodeCacheDeleteHashKey
	CodeCacheSetExpiration
	CodeCacheDecode
	CodeCacheLockNotAcquired
	CodeCacheLockFailed
	CodeCacheInvalidCastType
	CodeCacheNotFound
)
