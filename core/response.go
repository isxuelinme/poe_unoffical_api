package core

type GptMessage struct {
	Id              string   `json:"id"`
	Account         string   `json:"account"`
	ParentMessageId string   `json:"parent_message_id"`
	ConversationId  string   `json:"conversation_id"`
	IsEnd           bool     `json:"is_end"`
	Text            string   `json:"text"`
	Suggestions     []string `json:"suggestions"`
	Error           any      `json:"error"`
}

type CallbackMessageResponse struct {
	Type             string `json:"type"`
	CallbackFuncName string `json:"callback_func_name"`
	Data             any    `json:"data"`
}
