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
var NOT_AUTHORIZED = ResponseCode{Code: 3001, Message: "Not authorized!"}


var SYSTEM_ERROR = ResponseCode{Code: 5000, Message: "Server internal error!"}
var AUTH_FAILED = ResponseCode{Code: 5001, Message: "Authentication Failed!"}
var NOT_FOUND = ResponseCode{Code: 5002, Message: "Not found!"}


var INVALID_PARAMS = ResponseCode{Code: 4000, Message: "Invalid parameters!"}
var USER_ALREADY_REGISTERED = ResponseCode{Code: 4001, Message: "User already registered!"}
var UPLOAD_TYPE_NOT_SUPPORTED = ResponseCode{Code: 4002, Message: "Only support jpg or jpeg picture upload!"}
var CLUB_PIC_NUM_ABOVE_LIMIT = ResponseCode{Code: 4003, Message: "Club picture number above max limit!"}
var CLUB_TAG_NUM_ABOVE_LIMIT = ResponseCode{Code: 4004, Message: "Club tag number above max limit!"}
var INVALID_PICTURE_ID = ResponseCode{Code: 4005, Message: "Invalid picture id!"}
var PIC_TOO_LARGE = ResponseCode{Code: 4006, Message: "Picture too large MAX 1MB!"}
var WEB_SITE_TOO_LONG = ResponseCode{Code: 4007, Message: "Web site length above max limit 100 char!"}
var EMAIL_TOO_LONG = ResponseCode{Code: 4008, Message: "Email length above max limit 100 char!"}


func ConstructResponse(code ResponseCode, payload interface{}) Response {
	return Response{code, payload}
}

func SuccessResponse(payload interface{}) Response {
	return Response{ResponseCode: SUCCESS, Payload: payload}
}
