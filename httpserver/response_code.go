package httpserver

//ErrorCode definite

type ResponseCode struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

type Response struct {
	Status  ResponseCode `json:"status"`
	Payload interface{}  `json:"payload"`
}

var SUCCESS = ResponseCode{Code: 2000, Message: "Successful!"}
var AUTH_FAILED = ResponseCode{Code: 3001, Message: "Authentication Failed!"}

func ConstructResponse(code ResponseCode, payload interface{}) Response {
	return Response{code, payload}
}

func SuccessResponse(payload interface{}) Response {
	return Response{Status: SUCCESS, Payload: payload}
}
