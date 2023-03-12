package messageQueue

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"github.com/isxuelinme/poe_unoffical_api/internal/log"
	"sync"
	"time"
	//"xuelin.me/go/lib/log"
)

type RunMode uint8

const (
	ModeQueue RunMode = 1
	ModeChan  RunMode = 2
)

type MemoryMessageQueue struct {
	context                 context.Context
	cancel                  context.CancelFunc
	RunMode                 RunMode
	messageChan             chan *Message
	isPause                 bool
	pauseLock               *sync.Mutex
	Topic                   string
	List                    *list.List
	messageLock             *sync.Mutex
	Consumers               map[string]*ConsumerConfigure
	consumersLocker         *sync.Mutex
	ConsumerErrorCallback   Consumer
	ConsumerSuccessCallback Consumer
}

func NewMemoryMessageQueue(topic string, runMode RunMode) *MemoryMessageQueue {
	m := &MemoryMessageQueue{
		RunMode:         runMode,
		messageChan:     make(chan *Message),
		List:            list.New(),
		pauseLock:       &sync.Mutex{},
		messageLock:     &sync.Mutex{},
		consumersLocker: &sync.Mutex{},
		Consumers:       make(map[string]*ConsumerConfigure),
		Topic:           topic,
	}

	m.context, m.cancel = context.WithCancel(context.TODO())
	m.ConsumerErrorCallback = m.defaultConsumerErrorCallback
	m.ConsumerSuccessCallback = m.defaultConsumerSuccessCallback

	go m.work()
	return m
}
func (m *MemoryMessageQueue) SetConsumerErrorCallback(consumerErrorFunc Consumer) {
	m.ConsumerErrorCallback = consumerErrorFunc
}

func (m *MemoryMessageQueue) defaultConsumerErrorCallback(topic string, consumerId string, message *Message) error {
	return nil
}

func (m *MemoryMessageQueue) SetConsumerSuccessCallback(consumerSuccessFunc Consumer) {
	m.ConsumerSuccessCallback = consumerSuccessFunc
}

func (m *MemoryMessageQueue) defaultConsumerSuccessCallback(topic string, consumerId string, message *Message) error {
	return nil
}

func (m *MemoryMessageQueue) work() {
	if m.RunMode == ModeQueue {
		go m.workFromQueue()
	}
}

func (m *MemoryMessageQueue) callConsumer(topic string, consumerId string, consumerConfigure *ConsumerConfigure, msg *Message, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	err := consumerConfigure.Callback(m.Topic, consumerId, msg)
	if err != nil {
		if consumerConfigure.Retry == 0 {
			m.ConsumerErrorCallback(topic, consumerId, msg)
		} else {
			//如果消费出现错误执行三次，如果三次依然有问题就放弃执行，并将失败消息等写入执行失败队列
			for i := 0; i < consumerConfigure.Retry; i++ {
				err := consumerConfigure.Callback(m.Topic, consumerId, msg)
				if err == nil {
					break
				} else {
					if i == consumerConfigure.Retry-1 {
						//...对错误信息进行标记
						log.Error("try many times but still failed", err, consumerId, msg)
					}
				}

			}
		}
	} else {
		m.ConsumerSuccessCallback(topic, consumerId, msg)
	}
}

func (m *MemoryMessageQueue) workFromQueue() {
	for true {
		select {
		default:
			m.messageLock.Lock()
			msgElement := m.List.Front()
			if msgElement == nil {
				m.messageLock.Unlock()
				time.Sleep(time.Second * 3)
				continue
			}
			msg := msgElement.Value.(*Message)

			m.workByCall(msg)
			m.List.Remove(msgElement)
			m.messageLock.Unlock()
		case <-m.context.Done():
			return
		}
	}

}
func (m *MemoryMessageQueue) workByCall(message *Message) {
	select {
	default:
		if message.ConsumerId != "" {
			consumerConfigure, ok := m.Consumers[message.ConsumerId]
			if ok {
				consumerConfigure.Callback(m.Topic, consumerConfigure.Id, message)
			} else {
				log.Error("consumerId not found", message.ConsumerId)
			}
		} else {
			waitGroup := sync.WaitGroup{}
			waitGroup.Add(len(m.Consumers))
			for consumerId := range m.Consumers {
				go m.callConsumer(m.Topic, consumerId, m.Consumers[consumerId], message, &waitGroup)
			}
			waitGroup.Wait()
		}
	case <-m.context.Done():
		return
	}

}

type Publish struct {
	Message  *Message
	DisOrder bool
}

func (m *MemoryMessageQueue) Publish(message *Message) error {
	if m.isPause {
		if m.RunMode == ModeQueue {
			m.List.PushBack(message)
			return errors.New(fmt.Sprintf("Paused but you continute to publish,the message will be consumed in future : %s isPause: %v", m.Topic, m.isPause))
		}
		return errors.New(fmt.Sprintf("Paused Topic : %s isPause: %v", m.Topic, m.isPause))
	} else {
		if m.RunMode == ModeQueue {
			m.messageLock.Lock()
			defer m.messageLock.Unlock()
			m.List.PushBack(message)
		} else {
			if message.DisOrder {
				go m.workByCall(message)
			} else {
				m.messageLock.Lock()
				defer m.messageLock.Unlock()
				m.workByCall(message)
			}
		}

	}
	return nil
}

func (m *MemoryMessageQueue) AddConsumer(configure *ConsumerConfigure) {
	m.consumersLocker.Lock()
	defer m.consumersLocker.Unlock()
	m.Consumers[configure.Id] = configure
}

func (m *MemoryMessageQueue) RemoveConsumer(id string) {
	m.consumersLocker.Lock()
	defer m.consumersLocker.Unlock()
	delete(m.Consumers, id)
}

func (m *MemoryMessageQueue) GetConsumerCount() int {
	return len(m.Consumers)
}

func (m *MemoryMessageQueue) Pause() {
	m.pauseLock.Lock()
	defer m.pauseLock.Unlock()
	if m.isPause {
		return
	}
	m.isPause = true
	m.cancel()
}

// CancelPause resume
func (m *MemoryMessageQueue) CancelPause() {
	m.pauseLock.Lock()
	defer m.pauseLock.Unlock()
	if m.isPause == false {
		return
	}
	m.context, m.cancel = context.WithCancel(context.TODO())
	m.isPause = false
	m.work()
}
