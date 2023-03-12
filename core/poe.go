package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"poe_unoffical_api/internal/log"
	"poe_unoffical_api/internal/messageQueue"
	"sync"
	"time"
)

type PoeGPT struct {
	MessageBus                                 *messageQueue.MemoryMessageQueue
	PoeURl                                     string
	MessageID, ConversationID, ParentMessageID string
	wsOpen                                     bool
	wsReopenChan                               chan bool
	lastAskRequest                             *AskRequest
	cookie                                     string
	channel                                    string
	formKey                                    string
}

func NewPoeGPT(parentMessageID string, messageConsumer *messageQueue.ConsumerConfigure) *PoeGPT {
	poeGPT := &PoeGPT{
		MessageBus:      messageQueue.NewMemoryMessageQueue("Poe", messageQueue.ModeChan),
		PoeURl:          "https://Poe.khanh.lol/completion",
		ParentMessageID: parentMessageID,
		ConversationID:  "poe",
		wsOpen:          false,
		wsReopenChan:    make(chan bool),
	}

	cookie := os.Getenv("POE_COOKIE")
	if cookie == "" {
		log.PanicAndStopWorld("init POE configuration wrong, cookie is empty")
	}

	poeGPT.cookie = cookie
	_, _, _, _, err := poeGPT.GetSettings()
	if err != nil {
		log.PanicAndStopWorld("init POE configuration wrong ", err)
		return nil
	}

	poeGPT.MessageBus.AddConsumer(messageConsumer)
	go poeGPT.ReOpenWsClient()
	go poeGPT.PoeWsClient()
	time.Sleep(time.Second * 2) //waiting for websocket open
	return poeGPT
}

func (poe *PoeGPT) NewPoeRequest(parentMessageId, prompt string) *PoeRequest {
	return &PoeRequest{
		ParentMessageId: parentMessageId,
		Prompt:          prompt,
		Mode:            "Creative",
	}
}
func (poe *PoeGPT) GetMidCidAndPid() (string, string, string) {
	//TODO implement me
	return poe.MessageID, poe.ConversationID, poe.ParentMessageID
}
func (poe *PoeGPT) Publish(message *messageQueue.Message) {
	//TODO implement me
	poe.MessageBus.Publish(message)
}

func (poe *PoeGPT) AddConsumer(configure *messageQueue.ConsumerConfigure) {
	//TODO implement me
	poe.MessageBus.AddConsumer(configure)
}

func (poe *PoeGPT) RemoveConsumer(consumerName string) {
	poe.MessageBus.RemoveConsumer(consumerName)
}

func (poe *PoeGPT) GetMessageId() string {
	//TODO implement me
	return poe.MessageID
}

func (poe *PoeGPT) SetConversationID(cid string) {
	//TODO implement me
	poe.ConversationID = cid
}

func (poe *PoeGPT) SetParentMessageID(parentMessageID string) {
	//TODO implement me
	poe.ParentMessageID = parentMessageID
}

func (poe *PoeGPT) Talk(ctx context.Context, askRequest *AskRequest) (*GptMessage, error) {
	//TODO implement me

	type Variables struct {
		Bot           string      `json:"bot"`
		ChatID        int         `json:"chatId"`
		Query         string      `json:"query"`
		Source        interface{} `json:"source"`
		WithChatBreak bool        `json:"withChatBreak"`
	}
	type Payload struct {
		OperationName string    `json:"operationName"`
		Query         string    `json:"query"`
		Variables     Variables `json:"variables"`
	}
	payLodTemp := `
	{
  "operationName": "AddHumanMessageMutation",
  "query": "mutation AddHumanMessageMutation($chatId: BigInt!, $bot: String!, $query: String!, $source: MessageSource, $withChatBreak: Boolean! = false) {\n  messageEdgeCreate(\n    chatId: $chatId\n    bot: $bot\n    query: $query\n    source: $source\n    withChatBreak: $withChatBreak\n  ) {\n    __typename\n    message {\n      __typename\n      node {\n        __typename\n        ...MessageFragment\n        chat {\n          __typename\n          id\n          shouldShowDisclaimer\n        }\n      }\n    }\n    chatBreak {\n      __typename\n      node {\n        __typename\n        ...MessageFragment\n      }\n    }\n  }\n}\nfragment MessageFragment on Message {\n  id\n  __typename\n  messageId\n  text\n  linkifiedText\n  authorNickname\n  state\n  vote\n  voteReason\n  creationTime\n  suggestedReplies\n}",
  "variables": {
    "bot": "capybara",
    "chatId": 550922,
    "query": "现在还记得吗？\n",
    "source": null,
    "withChatBreak": false
  }
}
`

	var payload Payload

	json.Unmarshal([]byte(payLodTemp), &payload)

	payload.Variables.Query = askRequest.Question

	payloadBytes, err := json.Marshal(payload)
	body := bytes.NewReader(payloadBytes)
	resp, err := poe.Request(body)
	if err != nil {
		log.Error(err, "请求错误")
		poe.MessageBus.Publish(&messageQueue.Message{
			Id: time.Now().UnixNano(),
			MessageEntry: &GptMessage{
				IsEnd: true,
				Error: err.Error(),
			},
		})
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Error("statusCode is not 200")
		poe.MessageBus.Publish(&messageQueue.Message{
			Id: time.Now().UnixNano(),
			MessageEntry: &GptMessage{
				IsEnd: true,
				Error: "statusCode is not 200",
			},
		})
		return nil, err
	}

	defer resp.Body.Close()
	poe.lastAskRequest = askRequest

	if poe.wsOpen {
		return nil, nil
	} else {
		poe.wsReopenChan <- true
	}

	return nil, nil
	////http request
	//var poeHistoryNode PoeHistoryNode
	//for {
	//	resp, err := poe.GetHistory()
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		poe.MessageBus.Publish(&messageQueue.Message{
	//			Id: time.Now().UnixNano(),
	//			MessageEntry: &GptMessage{
	//				IsEnd: true,
	//				Error: err.Error(),
	//			},
	//		})
	//		return nil, err
	//	}
	//	nodeLen := len(resp.Data.ChatOfBot.MessagesConnection.Edges)
	//	if nodeLen > 0 && resp.Data.ChatOfBot.MessagesConnection.Edges[nodeLen-1].Node.State == "complete" {
	//		poeHistoryNode = resp.Data.ChatOfBot.MessagesConnection.Edges[nodeLen-1].Node
	//		break
	//	}
	//
	//	time.Sleep(time.Millisecond * 500)
	//}
	//
	//res := &GptMessage{
	//	Id:              fmt.Sprintf("%d", poeHistoryNode.MessageId),
	//	Account:         "1",
	//	IsEnd:           true,
	//	ConversationId:  "poe",
	//	Text:            poeHistoryNode.Text,
	//	ParentMessageId: poe.ParentMessageID,
	//}
	//
	//poe.MessageBus.Publish(&messageQueue.Message{
	//	Id:           time.Now().UnixNano(),
	//	MessageEntry: res,
	//})
	//
	//poe.ParentMessageID = fmt.Sprintf("%d", poeHistoryNode.MessageId)
	////temp 临时如此处理
	//poe.MessageID = fmt.Sprintf("%d", poeHistoryNode.MessageId)
	//poe.ConversationID = "poe"
	//
	//return res, nil
}

func (poe *PoeGPT) GetHistory() (*PoeGetHistoryMessage, error) {

	type Payload struct {
		OperationName string `json:"operationName"`
		Query         string `json:"query"`
		Variables     struct {
			Before interface{} `json:"before"`
			Bot    string      `json:"bot"`
			Last   int         `json:"last"`
		} `json:"variables"`
	}

	payLodTemp := `
	{
  "operationName": "ChatPaginationQuery",
  "query": "query ChatPaginationQuery($bot: String!, $before: String, $last: Int! = 10) {\n  chatOfBot(bot: $bot) {\n    id\n    __typename\n    messagesConnection(before: $before, last: $last) {\n      __typename\n      pageInfo {\n        __typename\n        hasPreviousPage\n      }\n      edges {\n        __typename\n        node {\n          __typename\n          ...MessageFragment\n        }\n      }\n    }\n  }\n}\nfragment MessageFragment on Message {\n  id\n  __typename\n  messageId\n  text\n  linkifiedText\n  authorNickname\n  state\n  vote\n  voteReason\n  creationTime\n  suggestedReplies\n}",
  "variables": {
    "before": null,
    "bot": "capybara",
    "last": 20
  }
}
`
	var payload Payload

	json.Unmarshal([]byte(payLodTemp), &payload)

	payload.Variables.Last = 1
	payloadBytes, _ := json.Marshal(payload)
	body := bytes.NewReader(payloadBytes)

	resp, err := poe.Request(body)
	if err != nil {
		log.Error(err, "请求错误")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Error("statusCode is not 200")
		return nil, errors.New("statusCode is not 200")
	}
	respB, _ := ioutil.ReadAll(resp.Body)

	respJson := PoeGetHistoryMessage{}

	json.Unmarshal(respB, &respJson)

	return &respJson, nil
}

type WsMessage struct {
	Messages []string `json:"messages"`
	MinSeq   int64    `json:"min_seq"`
}

type WsMessageStringToJson struct {
	MessageType string `json:"message_type"`
	Payload     struct {
		UniqueID         string `json:"unique_id"`
		SubscriptionName string `json:"subscription_name"`
		Data             struct {
			MessageAdded struct {
				ID               string      `json:"id"`
				MessageID        int         `json:"messageId"`
				CreationTime     int64       `json:"creationTime"`
				State            string      `json:"state"`
				Text             string      `json:"text"`
				Author           string      `json:"author"`
				LinkifiedText    string      `json:"linkifiedText"`
				SuggestedReplies []string    `json:"suggestedReplies"`
				Vote             interface{} `json:"vote"`
				VoteReason       interface{} `json:"voteReason"`
			} `json:"messageAdded"`
		} `json:"data"`
	} `json:"payload"`
}

var wsMessage *WsMessage
var wsMessageStringToJson *WsMessageStringToJson

func (poe *PoeGPT) GetSettings() (seq string, hash string, boxName string, channel string, err error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://poe.com/api/settings?channel=poe-chan51-8888-hhmpqzuksgonnzdwnitj", nil)
	req.Header.Set("authority", "poe.com")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("cookie", poe.cookie)
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("referer", "https://poe.com/sage")
	req.Header.Set("sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="100", "Google Chrome";v="100"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return "", "", "", "", err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return "", "", "", "", err
	}
	type GetSeqResponse struct {
		Formkey      string `json:"formkey"`
		TchannelData struct {
			MinSeq          string `json:"minSeq"`
			Channel         string `json:"channel"`
			ChannelHash     string `json:"channelHash"`
			BoxName         string `json:"boxName"`
			BaseHost        string `json:"baseHost"`
			TargetUrl       string `json:"targetUrl"`
			EnableWebsocket bool   `json:"enableWebsocket"`
		} `json:"tchannelData"`
	}
	var getSeqResponse GetSeqResponse
	json.Unmarshal(bodyText, &getSeqResponse)
	//to int64
	//strings.
	seq = getSeqResponse.TchannelData.MinSeq
	hash = getSeqResponse.TchannelData.ChannelHash
	boxName = getSeqResponse.TchannelData.BoxName
	channel = getSeqResponse.TchannelData.Channel
	poe.formKey = getSeqResponse.Formkey
	poe.channel = channel
	return
}

var littleLock = &sync.Mutex{}

func (poe *PoeGPT) PoeWsClient() bool {
	littleLock.Lock()
	defer littleLock.Unlock()
	if poe.wsOpen {
		log.Info("WebSocket is already open")
		return poe.wsOpen
	}
	log.Info("WebSocket is not open, start to open it")
	seq, hash, boxname, channel, err := poe.GetSettings()
	if err != nil {
		return false
	}
	wsMessage = &WsMessage{}
	wsMessageStringToJson = &WsMessageStringToJson{}
	url := fmt.Sprintf("wss://tch%d.tch.quora.com/up/%s/updates?min_seq=%s&channel=%s&hash=%s", rand.Intn(1000000)+1, boxname, seq, channel, hash)
	header := http.Header{}
	conn, resp, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		log.Error("dial:", err, resp.StatusCode)
		res := &GptMessage{
			IsEnd: true,
			Error: err.Error(),
		}
		poe.MessageBus.Publish(&messageQueue.Message{
			Id:           time.Now().UnixNano(),
			MessageEntry: res,
		})
		return false
	}
	poe.wsOpen = true
	defer conn.Close()
	// 在新协程中读取来自服务器的消息
	done := make(chan bool)
	go func() {
		lastMessage := ""
		lastMessageId := 0
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Error("websocket failed:", err)
				poe.wsOpen = false
				res := &GptMessage{
					IsEnd: true,
					Error: err.Error(),
				}
				poe.MessageBus.Publish(&messageQueue.Message{
					Id:           time.Now().UnixNano(),
					MessageEntry: res,
				})
				return
			}
			json.Unmarshal(message, wsMessage)
			json.Unmarshal([]byte(wsMessage.Messages[0]), wsMessageStringToJson)
			var res *GptMessage
			//sometimes the lastAskRequest is nil, because it is a pointer, set it in talk method
			//also we dont consumer the message when lastAskRequest is nil, because the message may be send by poe another service
			if poe.lastAskRequest == nil {
				continue
			} else {
				//you will poe your ask from websocket
				if wsMessageStringToJson.Payload.Data.MessageAdded.Text == poe.lastAskRequest.Question {
					continue
				}
				//when the lastMessage is not init value and the current message is the same as the last message, it means the message is not to send to consumer
				if lastMessage != "" && wsMessageStringToJson.Payload.Data.MessageAdded.Text == lastMessage {
					//but if the message is complete and the message id is not the same as the last save id, it means the message is to send to consumer
					if wsMessageStringToJson.Payload.Data.MessageAdded.State == "complete" && wsMessageStringToJson.Payload.Data.MessageAdded.MessageID != lastMessageId {
						lastMessageId = wsMessageStringToJson.Payload.Data.MessageAdded.MessageID
						res = &GptMessage{
							Id:              fmt.Sprintf("%d", wsMessageStringToJson.Payload.Data.MessageAdded.MessageID),
							Account:         "1",
							IsEnd:           true,
							ConversationId:  "poe",
							Text:            wsMessageStringToJson.Payload.Data.MessageAdded.Text,
							ParentMessageId: poe.ParentMessageID,
						}

						poe.MessageBus.Publish(&messageQueue.Message{
							Id:           time.Now().UnixNano(),
							MessageEntry: res,
						})

					}
					//even the message state is complete , but the suggestions after on complete, so if the suggestion is not empty, it means the message is to send to consumer
					if len(wsMessageStringToJson.Payload.Data.MessageAdded.SuggestedReplies) > 0 {
						res = &GptMessage{
							Id:              fmt.Sprintf("%d", wsMessageStringToJson.Payload.Data.MessageAdded.MessageID),
							Account:         "1",
							IsEnd:           true,
							ConversationId:  "poe",
							Text:            wsMessageStringToJson.Payload.Data.MessageAdded.Text,
							ParentMessageId: poe.ParentMessageID,
							Suggestions:     wsMessageStringToJson.Payload.Data.MessageAdded.SuggestedReplies,
						}
						poe.MessageBus.Publish(&messageQueue.Message{
							Id:           time.Now().UnixNano(),
							MessageEntry: res,
						})

					}
					continue
				}

			}

			lastMessage = wsMessageStringToJson.Payload.Data.MessageAdded.Text
			res = &GptMessage{
				Id:              fmt.Sprintf("%d", wsMessageStringToJson.Payload.Data.MessageAdded.MessageID),
				Account:         "1",
				IsEnd:           false,
				ConversationId:  "poe",
				Text:            wsMessageStringToJson.Payload.Data.MessageAdded.Text,
				ParentMessageId: poe.ParentMessageID,
			}
			poe.MessageBus.Publish(&messageQueue.Message{
				Id:           time.Now().UnixNano(),
				MessageEntry: res,
			})
			poe.ParentMessageID = fmt.Sprintf("%d", wsMessageStringToJson.Payload.Data.MessageAdded.MessageID)
			poe.MessageID = fmt.Sprintf("%d", wsMessageStringToJson.Payload.Data.MessageAdded.MessageID)
			poe.ConversationID = "poe"
			//fmt.Printf("message_id:%d\ncontent:%s\ncomplete%s", wsMessageStringToJson.Payload.Data.MessageAdded.MessageID, wsMessageStringToJson.Payload.Data.MessageAdded.Text, wsMessageStringToJson.Payload.Data.MessageAdded.State)
		}
	}()

	go func() {
		defer func() {
			done <- true
		}()
		for {
			err := conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
			if err != nil {
				log.Error("websocket failed:", err)
				return
			}
			time.Sleep(time.Second * 3)
		}
	}()

	select {
	case <-done:
		log.Info("\tcase <-done:\n")
		poe.wsOpen = false
		poe.wsReopenChan <- true
		log.Error("websocket failed reopen:")
		close(done)
		return false
	}
}

func (poe *PoeGPT) ReOpenWsClient() {
	for {
		select {
		case <-poe.wsReopenChan:
			log.Info("websocket reopen to reconnect")
			go poe.PoeWsClient()
		}
	}

}

func (poe *PoeGPT) Request(body *bytes.Reader) (*http.Response, error) {
	url := "https://poe.com/api/gql_POST"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("authority", "poe.com")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", poe.cookie)
	req.Header.Set("origin", "https://poe.com")
	req.Header.Set("poe-formkey", poe.formKey)
	req.Header.Set("poe-tchannel", poe.channel)
	req.Header.Set("referer", "https://poe.com/")
	req.Header.Set("sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="100", "Google Chrome";v="100"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
