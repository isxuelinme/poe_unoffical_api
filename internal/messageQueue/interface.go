package messageQueue

type Message struct {
	Id           any
	MessageEntry interface{}
	DisOrder     bool   //if you set this true, the message will be consumed as soon as possible but not in order
	ConsumerId   string //if not set,push to all consumer

}
type Consumer func(topic string, string string, message *Message) error
type ConsumerConfigure struct {
	Id       string
	Retry    int
	Callback Consumer
}
