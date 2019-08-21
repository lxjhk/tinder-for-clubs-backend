package httpserver

//ErrorCode definite

type ResponseCode struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

type Response struct {
	ResponseCode
	Payload interface{}  `json:"payload"`
}

var SUCCESS = ResponseCode{Code: 2000, Message: "Successful!"}

var NO_PERMISSION = ResponseCode{Code: 3000, Message: "No permission!"}

var SYSTEM_ERROR = ResponseCode{Code: 5000, Message: "Server internal error!"}
var AUTH_FAILED = ResponseCode{Code: 5001, Message: "Authentication Failed!"}
var SAVE_SESSION_FAILED = ResponseCode{Code: 5002, Message: "Failed to save session!"}
var SAVE_PICTURE_FAILED = ResponseCode{Code: 5002, Message: "Failed to save picture!"}


var INVALID_PARAMS = ResponseCode{Code: 4000, Message: "Invalid parameters!"}
var PIC_NUM_NOT_SUPPORTED = ResponseCode{Code: 4001, Message: "Picture number uploaded not supported!"}
var UPLOAD_TYPE_NOT_SUPPORTED = ResponseCode{Code: 4001, Message: "Only support picture upload!"}
var NOT_FOUND = ResponseCode{Code: 4001, Message: "Data not found!"}

func ConstructResponse(code ResponseCode, payload interface{}) Response {
	return Response{code, payload}
}

func SuccessResponse(payload interface{}) Response {
	return Response{ResponseCode: SUCCESS, Payload: payload}
}
