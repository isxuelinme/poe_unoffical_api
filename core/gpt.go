package core

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"poe_unoffical_api/internal/log"
	"poe_unoffical_api/internal/messageQueue"
	"time"
	//"xuelin.me/go/lib/log"
)

var multiUserGpt *MultiUserGpt

type GptInterface interface {
	Talk(ctx context.Context, askRequest *AskRequest) (*GptMessage, error)
	Publish(message *messageQueue.Message)
	AddConsumer(consumer *messageQueue.ConsumerConfigure)
	RemoveConsumer(consumerName string)
	GetMessageId() string
	SetConversationID(cid string)
	SetParentMessageID(parentMessageID string)
	GetMidCidAndPid() (mid, cid, pid string)
}
type GptType uint

var GptTypeBingGpt GptType = 1
var GptTypeUnOfficial GptType = 2
var GptTypeOfficialApi GptType = 3
var GptTypePoeUnofficial GptType = 4

type UserChat struct {
	UserId        int64
	Gpt           GptInterface
	lastMessageId string
}
type MultiUserGpt struct {
	gptType GptType //1:bing 2:unofficial 3:official
	GptMap  map[int64]*UserChat
}

func NewMutLtiUserGpt(gptType GptType) *MultiUserGpt {
	if multiUserGpt != nil {
		return multiUserGpt
	}
	multiUserGpt = &MultiUserGpt{
		gptType: gptType,
		GptMap:  make(map[int64]*UserChat),
	}
	return multiUserGpt
}

func (m *MultiUserGpt) Talk(askRequest *AskRequest) {
	if userChat, ok := m.GptMap[askRequest.UserId]; ok {
		userChat.Ask(askRequest)
	} else {
		m.GptMap[askRequest.UserId] = &UserChat{
			UserId: askRequest.UserId,
		}

		//d := &db.ChatMessages{}

		if m.gptType == GptTypeBingGpt {
			//cid, pid, _ := d.GetLastConversationIdAndMessageIdByUserId(1, "bing")
			var cid, pid string = "1", "2"

			userStore := &UserStore{
				UserId:          1,
				ConversationId:  cid,
				ParentMessageId: pid,
			}
			m.GptMap[askRequest.UserId].NewBingGpt(userStore)
		} else if m.gptType == GptTypeUnOfficial {
			//cid, pid, _ := d.GetLastConversationIdAndMessageIdByUserId(1, "unofficial")
			var cid, pid string = "1", "2"

			userStore := &UserStore{
				UserId:          1,
				ConversationId:  cid,
				ParentMessageId: pid,
			}
			m.GptMap[askRequest.UserId].NewUnofficialChatGpt(userStore)
		} else if m.gptType == GptTypePoeUnofficial {
			//cid, pid, _ := d.GetLastConversationIdAndMessageIdByUserId(1, "poe")
			var cid, pid string = "1", "2"

			userStore := &UserStore{
				UserId:          1,
				ConversationId:  cid,
				ParentMessageId: pid,
			}
			m.GptMap[askRequest.UserId].NewPoeGpt(userStore)
		}
		m.GptMap[askRequest.UserId].Ask(askRequest)
	}
}

func (u *UserChat) Ask(askRequest *AskRequest) {
	//var lastMessage = ""
	//var lastSecTimeStamp = time.Now().Unix()
	var lastResponse *GptMessage
	ctx, cancel := context.WithCancel(context.Background())
	u.Gpt.RemoveConsumer("messageNotifyBus")
	u.Gpt.AddConsumer(u.NewMessageNotify(askRequest, lastResponse, ctx, cancel))
	u.Gpt.Talk(ctx, askRequest)
}

func (u *UserChat) NewMessageNotify(askRequest *AskRequest, lastResponse *GptMessage, ctx context.Context, cancel context.CancelFunc) *messageQueue.ConsumerConfigure {

	var lastMessage = ""
	var lastMessageTimeStamp = time.Now().Unix()

	go func(cancelFunc context.CancelFunc) {
		for {
			select {
			case <-ctx.Done():
				log.Info("Cancel", ctx.Err())
				return
			default:
				if time.Now().Unix()-lastMessageTimeStamp > 60 {
					if lastResponse != nil {
						lastResponse.IsEnd = true
						lastResponse.Error = "timeout"
						u.Gpt.Publish(&messageQueue.Message{
							Id:           lastResponse.Id,
							MessageEntry: lastResponse,
						})
					} else {
						lastResponse = &GptMessage{
							IsEnd: true,
							Error: "timeout",
						}
						u.Gpt.Publish(&messageQueue.Message{
							Id:           time.Now().UnixNano(),
							MessageEntry: lastResponse,
						})
					}
					fmt.Println("cancel timeout")
					cancelFunc()
					return
				}
			}
			time.Sleep(time.Second)
		}
	}(cancel)

	consumerBus := messageQueue.ConsumerConfigure{
		Id:    "messageNotifyBus",
		Retry: 0,
		Callback: func(topic string, consumerId string, message *messageQueue.Message) error {
			lastMessageTimeStamp = time.Now().Unix()
			orgMessage := message.MessageEntry.(*GptMessage)
			if orgMessage.IsEnd {
				dialCallBackResp := &GptMessage{}
				if orgMessage.Text == "" {
					dialCallBackResp = &GptMessage{
						IsEnd:          true,
						Id:             orgMessage.Id,
						ConversationId: orgMessage.ConversationId,
						Text:           "",
						Error:          orgMessage.Error,
						Suggestions:    orgMessage.Suggestions,
					}
				} else {
					dialCallBackResp = &GptMessage{
						IsEnd:          true,
						Id:             orgMessage.Id,
						ConversationId: orgMessage.ConversationId,
						Text:           orgMessage.Text,
						Error:          orgMessage.Error,
						Suggestions:    orgMessage.Suggestions,
					}
				}
				sendData, _ := json.Marshal(WSMessageResponse{
					Type:             "dialServiceWithCallbackResponse",
					CallbackFuncName: askRequest.CallbackFuncName,
					Data:             dialCallBackResp,
				})
				//ws.GetManager().SendAll(sendData)
				askRequest.AskResponseCallBack(askRequest, sendData)
				if orgMessage.Text == "" {
					log.Error("发生错误", orgMessage.Error)
				}
				cancel()
			} else {

				if lastMessage == "" {
					dialCallBackResp := &GptMessage{
						Id:             orgMessage.Id,
						ConversationId: orgMessage.ConversationId,
						Text:           " ",
						Suggestions:    []string{},
					}
					sendData, _ := json.Marshal(WSMessageResponse{
						Type:             "dialServiceWithCallbackResponse",
						CallbackFuncName: askRequest.CallbackFuncName,
						Data:             dialCallBackResp,
					})
					//ws.GetManager().SendAll(sendData)
					askRequest.AskResponseCallBack(askRequest, sendData)
				}

				if lastMessage == orgMessage.Text && len(orgMessage.Suggestions) == 0 {
					return nil
				} else {
					lastMessage = orgMessage.Text
					var suggestions []string
					if len(orgMessage.Suggestions) == 0 {
						suggestions = []string{}
					} else {
						suggestions = orgMessage.Suggestions
					}
					dialCallBackResp := &GptMessage{
						Id:             orgMessage.Id,
						ConversationId: orgMessage.ConversationId,
						Text:           orgMessage.Text,
						Suggestions:    suggestions,
					}
					sendData, _ := json.Marshal(WSMessageResponse{
						Type:             "dialServiceWithCallbackResponse",
						CallbackFuncName: askRequest.CallbackFuncName,
						Data:             dialCallBackResp,
					})
					//ws.GetManager().SendAll(sendData)
					askRequest.AskResponseCallBack(askRequest, sendData)
					fmt.Printf("receive %v topic:  %s consumerId %s time %s \n", message.Id, topic, consumerId, time.Now().Format("2006-01-02 15:04:05"))
				}
			}
			return nil
		},
	}

	return &consumerBus
}

type UserStore struct {
	UserId          int64
	ConversationId  string
	ParentMessageId string
}

func (u *UserChat) NewUnofficialChatGpt(store *UserStore) {
	//implement it by yourself :)
}

func (u *UserChat) NewBingGpt(store *UserStore) {
	//implement it by yourself :)

}
func (u *UserChat) NewPoeGpt(store *UserStore) {

	if u.Gpt != nil {
		return
	}
	u.Gpt = NewPoeGPT(store.ParentMessageId, u.NewPersistenceConsumer())

	u.Gpt.SetConversationID("poe")

	u.Gpt.SetParentMessageID(store.ParentMessageId)

}

func (u *UserChat) NewPersistenceConsumer() *messageQueue.ConsumerConfigure {
	DbInsertConsumer := &messageQueue.ConsumerConfigure{
		Id:    "dbInsert",
		Retry: 0,
		Callback: func(topic string, consumerId string, message *messageQueue.Message) error {
			orgMessage := message.MessageEntry.(*GptMessage)
			if orgMessage.IsEnd {
				if orgMessage.Id == u.lastMessageId {
					log.Info("db_insert repeat message id")
					return nil
				}
				u.lastMessageId = orgMessage.Id
				log.Info("db_insert", orgMessage)
				if orgMessage.Text == "" {
					log.Error("response wrong ", orgMessage.Error)
					return nil
				}
				// e.g you also do it
				//insert := &db.ChatMessages{
				//	UserID:          1,
				//	AccountNumber:   "1",
				//	MessageID:       orgMessage.Id,
				//	Type:            2,
				//	ConversationID:  orgMessage.ConversationId,
				//	ParentMessageID: u.Gpt.GetMessageId(),
				//	Contents:        orgMessage.Text,
				//	CreatedAt:       time.Now().UnixNano(),
				//}
				//_, err := insert.Add()
				//if err != nil {
				//	log.Error("insert db failed", err)
				//}
			}
			return nil
		},
	}
	return DbInsertConsumer
}
