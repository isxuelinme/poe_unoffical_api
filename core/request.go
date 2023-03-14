package core

type AskRequestCallBack func(askRequest *AskRequest, response []byte)

// AskRequest like the JSONP callback if you are older ,you know what i mean:)
// CallbackFuncName when your client send it to server, also reply back to client
// when you are using websocket.Send, you can build a global callback function to handle the response
type AskRequest struct {
	UserId              int64
	Question            string
	CallbackFuncName    string
	AskResponseCallBack AskRequestResponseCallBack
}

type RequestConnType int

var RequestConnTypeWebsocket RequestConnType = 1
var RequestConnTypeJSON RequestConnType = 2
var RequestConnTypeGrpc RequestConnType = 3
var RequestConnTypeHttpEventStream = 4

type AskRequestResponseCallBack func(askRequest *AskRequest, message *CallbackMessageResponse)
