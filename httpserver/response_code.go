package httpserver

//ErrorCode definite

type ResponseCode struct {
	Code    int
	Message string
}

type Response struct {
	Status  ResponseCode
	Payload interface{}
}

var SUCCESS = ResponseCode{Code: 2000, Message: "Successful!"}
var AUTH_FAILED = ResponseCode{Code: 3001, Message: "Authentication Failed!"}

func ConstructResponse(code ResponseCode, payload interface{}) Response {
	return Response{code, payload}
}

func SuccessResponse(payload interface{}) Response {
	return Response{Status: SUCCESS, Payload: payload}
}
