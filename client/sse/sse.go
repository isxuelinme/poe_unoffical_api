package sse

import (
	"encoding/json"
	"github.com/alexandrevicenzi/go-sse"
	"github.com/isxuelinme/poe_unoffical_api/core"
	"net/http"
	"os"
)

type AskRequest struct {
	Type            string `json:"type"`
	ConversationId  string `json:"conversation_id"`
	ParentMessageId string `json:"parent_message_id"`
	Text            string `json:"text"`
}

type AskResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func sendMessage(mutilUser *core.MultiUserGpt, s *sse.Server, text string) {
	mutilUser.Talk(&core.AskRequest{
		UserId:           1,
		Question:         text,
		CallbackFuncName: "",
		AskResponseCallBack: func(askRequest *core.AskRequest, message *core.CallbackMessageResponse) {
			response, _ := json.Marshal(message)
			if message.Data.IsEnd {
				s.SendMessage("/events", sse.NewMessage("", string(response), "done"))
			} else {
				s.SendMessage("/events", sse.SimpleMessage(message.Data.Text))
			}
		},
	})
}

var MutLtiUserGpt *core.MultiUserGpt
var sseServer *sse.Server = nil

func SSE() {
	core.SetLogMode(core.LOG_ERROR)

	MutLtiUserGpt = core.NewMutLtiUserGpt(core.GptTypePoeUnofficial)
	// Create SSE server
	sseServer = sse.NewServer(nil)
	defer sseServer.Shutdown()

	// Configure the route
	http.Handle("/events", sseServer)
	http.Handle("/", http.FileServer(http.Dir("./public/")))
	http.Handle("/conversation", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a new AskRequest
		askRequest := &AskRequest{}
		// Decode the request body into the struct
		err := json.NewDecoder(r.Body).Decode(askRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(200)
		bytes, _ := json.Marshal(&AskResponse{
			Code: 200,
			Msg:  "ok",
		})
		w.Write(bytes)
		// Send the message to the SSE server
		sendMessage(MutLtiUserGpt, sseServer, askRequest.Text)
	}))
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8090"
	}
	http.ListenAndServe(":"+port, nil)
}
