package errors

import "net/http"

var ErrorMessages = ErrorMessage{
	CodeHTTPBadRequest:          ErrMsgBadRequest,
	CodeHTTPNotFound:            ErrMsgNotFound,
	CodeHTTPUnauthorized:        ErrMsgUnauthorized,
	CodeHTTPInternalServerError: ErrMsgISE,
	CodeHTTPUnmarshal:           ErrMsgBadRequest,
	CodeHTTPMarshal:             ErrMsgISE,
	CodeHTTPConflict:            ErrMsgConflict,
	CodeHTTPForbidden:           ErrMsgForbidden,
	CodeHTTPUnprocessableEntity: ErrMsgUnprocessable,
	CodeHTTPTooManyRequest:      ErrMsgTooManyRequest,
	CodeHTTPServiceUnavailable:  ErrMsgServiceUnavailable,
	CodeHTTPParamDecode:         ErrMsgBadRequest,
	CodeHTTPErrorOnReadBody:     ErrMsgISE,

	CodeSQLBuilder:                    ErrMsgISE,
	CodeSQLRead:                       ErrMsgISE,
	CodeSQLRowScan:                    ErrMsgISE,
	CodeSQLCreate:                     ErrMsgISE,
	CodeSQLUpdate:                     ErrMsgISE,
	CodeSQLDelete:                     ErrMsgISE,
	CodeSQLUnlink:                     ErrMsgISE,
	CodeSQLTxBegin:                    ErrMsgISE,
	CodeSQLTxCommit:                   ErrMsgISE,
	CodeSQLPrepareStmt:                ErrMsgISE,
	CodeSQLRecordMustExist:            ErrMsgNotFound,
	CodeSQLCannotRetrieveLastInsertID: ErrMsgISE,
	CodeSQLCannotRetrieveAffectedRows: ErrMsgISE,
	CodeSQLUniqueConstraint:           ErrMsgUniqueConst,
	CodeSQLRecordDoesNotMatch:         ErrMsgBadRequest,
	CodeSQLRecordIsExpired:            ErrMsgBadRequest,
	CodeSQLRecordDoesNotExist:         ErrMsgNotFound,
	CodeSQLForeignKeyMissing:          ErrMsgISE,

	CodeTokenStillValid:        ErrMsgTokenStillValid,
	CodeTokenRefreshStillValid: ErrMsgRefreshStillValid,
}

var (
	ErrMsgBadRequest = Message{
		StatusCode: http.StatusBadRequest,
		EN:         `Invalid Input. Please Validate Your Input.`,
		ID:         `Kesalahan Input. Mohon Cek Kembali Masukkan Anda.`,
	}
	ErrMsgNotFound = Message{
		StatusCode: http.StatusNotFound,
		EN:         `Record Does Not Exist. Please Validate Your Input Or Contact Administrator.`,
		ID:         `Data Tidak Ditemukan. Mohon Cek Kembali Masukkan Anda Atau Hubungi Administrator.`,
	}
	ErrMsgUnauthorized = Message{
		StatusCode: http.StatusUnauthorized,
		EN:         `Unauthorized Access. You are not authorized to access this resource.`,
		ID:         `Akses Ditolak. Anda Belum Diijinkan Untuk Mengakses Aplikasi.`,
	}
	ErrMsgISE = Message{
		StatusCode: http.StatusInternalServerError,
		EN:         `Internal Server Error. Please Call Administrator.`,
		ID:         `Terjadi Kendala Pada Server. Mohon Hubungi Administrator.`,
	}
	ErrMsgConflict = Message{
		StatusCode: http.StatusConflict,
		EN:         `Record has existed and must be unique. Please Validate Your Input Or Contact Administrator.`,
		ID:         `Data sudah ada. Mohon Cek Kembali Masukkan Anda Atau Hubungi Administrator.`,
	}
	ErrMsgForbidden = Message{
		StatusCode: http.StatusForbidden,
		EN:         `Forbidden. You don't have permission to access this resource.`,
		ID:         `Terlarang. Anda tidak memiliki izin untuk mengakses aplikasi.`,
	}
	ErrMsgUnprocessable = Message{
		StatusCode: http.StatusUnprocessableEntity,
		EN:         `Not Able to Process This Entity. Please Validate Your Input and Try Again`,
		ID:         `Entitas Ini Tidak Dapat Diproses. Mohon Cek Kembali Masukkan Anda Dan Coba Kembali`,
	}
	ErrMsgTooManyRequest = Message{
		StatusCode: http.StatusTooManyRequests,
		EN:         `Too Many Request For This Entity. Please Wait And Try Again.`,
		ID:         `Permintaan Terlalu Banyak Untuk Entitas Ini. Mohon Tunggu Dan Coba Kembali.`,
	}
	ErrMsgServiceUnavailable = Message{
		StatusCode: http.StatusServiceUnavailable,
		EN:         `Service is unavailable.`,
		ID:         `Layanan sedang tidak tersedia.`,
	}
	ErrMsgUniqueConst = Message{
		StatusCode: http.StatusConflict,
		EN:         `Record Has Existed and Must Be Unique. Please Validate Your Input Or Contact Administrator.`,
		ID:         `Data sudah ada. Mohon Cek Kembali Masukkan Anda Atau Hubungi Administrator.`,
	}
	ErrMsgTokenStillValid = Message{
		StatusCode: http.StatusForbidden,
		EN:         `Token still valid. Please Validate Your Input Or Contact Administrator.`,
		ID:         `Token masih valid. Mohon Cek Kembali Masukkan Anda Atau Hubungi Administrator.`,
	}
	ErrMsgRefreshStillValid = Message{
		StatusCode: http.StatusForbidden,
		EN:         `Refresh token still valid. Please Validate Your Input Or Contact Administrator.`,
		ID:         `Refresh token masih valid. Mohon Cek Kembali Masukkan Anda Atau Hubungi Administrator.`,
	}
)
