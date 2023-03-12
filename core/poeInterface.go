package core

type PoeRequest struct {
	Prompt          string `json:"prompt"`
	ParentMessageId string `json:"parentMessageId"`
	Mode            string `json:"mode"`
}

type PoeResponse struct {
	Response  string `json:"response"`
	MessageId string `json:"messageId"`
	Result    struct {
		Value   string      `json:"value"`
		Code    int         `json:"code"`
		Message interface{} `json:"message"`
	} `json:"result"`
}
type PoeSendMessageResponse struct {
	Data struct {
		MessageEdgeCreate struct {
			Typename string `json:"__typename"`
			Message  struct {
				Typename string `json:"__typename"`
				Node     struct {
					Typename         string      `json:"__typename"`
					Id               string      `json:"id"`
					MessageId        int         `json:"messageId"`
					Text             string      `json:"text"`
					LinkifiedText    string      `json:"linkifiedText"`
					AuthorNickname   string      `json:"authorNickname"`
					State            string      `json:"state"`
					Vote             interface{} `json:"vote"`
					VoteReason       interface{} `json:"voteReason"`
					CreationTime     int64       `json:"creationTime"`
					SuggestedReplies []string    `json:"suggestedReplies"`
					Chat             struct {
						Typename             string `json:"__typename"`
						Id                   string `json:"id"`
						ShouldShowDisclaimer bool   `json:"shouldShowDisclaimer"`
					} `json:"chat"`
				} `json:"node"`
			} `json:"message"`
			ChatBreak interface{} `json:"chatBreak"`
		} `json:"messageEdgeCreate"`
	} `json:"data"`
	Extensions struct {
		IsFinal bool `json:"is_final"`
	} `json:"extensions"`
}
type PoeHistoryNode struct {
	Typename         string      `json:"__typename"`
	Id               string      `json:"id"`
	MessageId        int         `json:"messageId"`
	Text             string      `json:"text"`
	LinkifiedText    string      `json:"linkifiedText"`
	AuthorNickname   string      `json:"authorNickname"`
	State            string      `json:"state"`
	Vote             interface{} `json:"vote"`
	VoteReason       interface{} `json:"voteReason"`
	CreationTime     int64       `json:"creationTime"`
	SuggestedReplies []string    `json:"suggestedReplies"`
}
type PoeGetHistoryMessage struct {
	Data struct {
		ChatOfBot struct {
			Id                 string `json:"id"`
			Typename           string `json:"__typename"`
			MessagesConnection struct {
				Typename string `json:"__typename"`
				PageInfo struct {
					Typename        string `json:"__typename"`
					HasPreviousPage bool   `json:"hasPreviousPage"`
				} `json:"pageInfo"`
				Edges []struct {
					Typename string         `json:"__typename"`
					Node     PoeHistoryNode `json:"node"`
				} `json:"edges"`
			} `json:"messagesConnection"`
		} `json:"chatOfBot"`
	} `json:"data"`
	Extensions struct {
		IsFinal bool `json:"is_final"`
	} `json:"extensions"`
}
