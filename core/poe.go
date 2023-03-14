package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/isxuelinme/poe_unoffical_api/internal/log"
	"github.com/isxuelinme/poe_unoffical_api/internal/messageQueue"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
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
	cookieMap                                  map[string]string
	setting                                    *GetSettingResponse
}

func NewPoeGPT(parentMessageID string, messageConsumer *messageQueue.ConsumerConfigure) *PoeGPT {
	log.Debug("start poe gpt")

	poeGPT := &PoeGPT{
		MessageBus:      messageQueue.NewMemoryMessageQueue("Poe", messageQueue.ModeChan),
		ParentMessageID: parentMessageID,
		ConversationID:  "poe",
		wsOpen:          false,
		wsReopenChan:    make(chan bool),
		cookieMap:       make(map[string]string),
	}

	cookie := os.Getenv("POE_COOKIE")
	if cookie == "" {
		log.PanicAndStopWorld("init POE configuration wrong, cookie is empty")
	}
	poeGPT.cookie = cookie

	poeChannel := os.Getenv("POE_CHANNEL")
	if poeChannel == "" {
		log.PanicAndStopWorld("init POE configuration wrong, cookie is empty")
	}

	_, _, _, _, err := poeGPT.GetSettings()
	if err != nil {
		log.PanicAndStopWorld("init setting wrong ", err)
		return nil
	}
	log.Debug("get settings success")

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
	log.Debug("new question:" + askRequest.Question)
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

	var payload Payload

	json.Unmarshal([]byte(payLoadForTalk), &payload)

	payload.Variables.ChatID, _ = strconv.Atoi(os.Getenv("POV_CHAT_ID"))
	payload.Variables.Query = askRequest.Question

	payloadBytes, err := json.Marshal(payload)
	body := bytes.NewReader(payloadBytes)
	url := "https://poe.com/api/gql_POST"
	resp, err := poe.PoeRequest("POST", url, body)
	if err != nil {
		log.Error("http request error", err)
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
		str, _ := io.ReadAll(resp.Body)
		log.Error("statusCode is not 200", resp.StatusCode, string(str))
		poe.MessageBus.Publish(&messageQueue.Message{
			Id: time.Now().UnixNano(),
			MessageEntry: &GptMessage{
				IsEnd: true,
				Error: "statusCode is not 200",
			},
		})
		return nil, err
	}

	str, _ := ioutil.ReadAll(resp.Body)

	log.Debug("answer:", string(str))
	defer resp.Body.Close()
	poe.lastAskRequest = askRequest

	if poe.wsOpen {
		return nil, nil
	} else {
		poe.wsReopenChan <- true
	}

	return nil, nil
	//although we can poll http request to get the response, but it respond till got entire response
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
	//poe.MessageID = fmt.Sprintf("%d", poeHistoryNode.MessageId)
	//poe.ConversationID = "poe"
	//
	//return res, nil
}

func (poe *PoeGPT) GetHistory() (*PoeGetHistoryMessage, error) {
	payload := PoeGetHistoryPayload{}
	json.Unmarshal([]byte(payLoadForGetHistory), &payload)

	payload.Variables.Last = 1
	payloadBytes, _ := json.Marshal(payload)
	body := bytes.NewReader(payloadBytes)
	url := "https://poe.com/api/gql_POST"
	resp, err := poe.PoeRequest("POST", url, body)
	if err != nil {
		log.Error("request err", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Error("statusCode is not 200", resp.StatusCode)
		return nil, errors.New("statusCode is not 200")
	}
	respB, _ := ioutil.ReadAll(resp.Body)

	respJson := PoeGetHistoryMessage{}

	err = json.Unmarshal(respB, &respJson)
	if err != nil {
		log.Error("json unmarshal error", err)
		return nil, err
	}

	return &respJson, nil
}

var wsMessage *poeWsMessage
var wsMessageStringToJson *poeWsMessageStringToJson

func (poe *PoeGPT) GetSettings() (seq string, hash string, boxName string, channel string, err error) {
	url := fmt.Sprintf("https://poe.com/api/settings?channel=%s", os.Getenv("POE_CHANNEL"))
	resp, err := poe.PoeRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
		return "", "", "", "", err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", "", err
	}
	var getSeqResponse GetSettingResponse
	err = json.Unmarshal(bodyText, &getSeqResponse)
	if err != nil {
		return "", "", "", "", err
	}
	//to int64
	//strings.
	seq = getSeqResponse.TchannelData.MinSeq
	hash = getSeqResponse.TchannelData.ChannelHash
	boxName = getSeqResponse.TchannelData.BoxName
	channel = getSeqResponse.TchannelData.Channel
	poe.setting = &getSeqResponse
	return
}

var littleLock = &sync.Mutex{}

func (poe *PoeGPT) PoeWsClient() bool {
	littleLock.Lock()
	defer littleLock.Unlock()
	if poe.wsOpen {
		log.Debug("WebSocket is already open")
		return poe.wsOpen
	}
	log.Debug("WebSocket is not open, start to open it")
	wsMessage = &poeWsMessage{}
	wsMessageStringToJson = &poeWsMessageStringToJson{}
	url := fmt.Sprintf("wss://tch%d.tch.quora.com/up/%s/updates?min_seq=%s&channel=%s&hash=%s", rand.Intn(1000000)+1, poe.setting.TchannelData.BoxName, poe.setting.TchannelData.MinSeq, poe.setting.TchannelData.Channel, poe.setting.TchannelData.ChannelHash)
	header := http.Header{}
	log.Debug("websocket url:", url)
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
			log.Debug("message:", string(message))
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
			log.Debug(fmt.Sprintf("message_id:%d\ncontent:%s\ncomplete%s", wsMessageStringToJson.Payload.Data.MessageAdded.MessageID, wsMessageStringToJson.Payload.Data.MessageAdded.Text, wsMessageStringToJson.Payload.Data.MessageAdded.State))

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
		log.Debug("case <-done:")
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
			log.Debug("websocket reopen to reconnect")
			go poe.PoeWsClient()
		}
	}

}

func (poe *PoeGPT) PoeRequest(httpMethod, url string, body *bytes.Reader) (*http.Response, error) {
	var req *http.Request
	var err error
	if body == nil {
		req, err = http.NewRequest(httpMethod, url, nil)
	} else {
		req, err = http.NewRequest(httpMethod, url, body)
	}
	if err != nil {
		return nil, err
	}
	headers := http.Header{
		"authority":          {"poe.com"},
		"accept":             {"*/*"},
		"accept-language":    {"zh-CN,zh;q=0.9"},
		"content-type":       {"application/json"},
		"cookie":             {poe.cookie},
		"origin":             {"https://poe.com"},
		"referer":            {"https://poe.com/"},
		"sec-ch-ua":          {`" Not A;Brand";v="99", "Chromium";v="100", "Google Chrome";v="100"`},
		"sec-ch-ua-mobile":   {"?0"},
		"sec-ch-ua-platform": {`"macOS"`},
		"sec-fetch-dest":     {"empty"},
		"sec-fetch-mode":     {"cors"},
		"sec-fetch-site":     {"same-origin"},
		"user-agent":         {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36"},
	}
	if url == "https://poe.com/api/gql_POST" {
		headers.Add("poe-formkey", poe.setting.Formkey)
		headers.Add("poe-tchannel", poe.setting.TchannelData.Channel)
	}
	req.Header = headers

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
