package message

type DelayMessage struct {
	// The topic of your message
	Topic string `json:"topic"`
	// Your message payload with base64 encoding
	Payload string `json:"payload"`
	// specify when to send your message relative to the current time
	ProcessIn int32 `json:"processIn"`
	// specify when to send your message ,if ProcessIn is not set
	ProcessAt int64 `json:"processAt"`
	// ToleranceSecond, default is 0
	//	- If messageReceivedTime.Before(messageProcessTime)
	//		then enqueue the message.
	//	- If messageReceivedTime.After(messageProcessTime) and messageReceivedTime.Sub(messageProcessTime) > ToleranceSecond
	//		then ignore the message.
	//	- If messageReceivedTime.After(messageProcessTime) and messageReceivedTime.Sub(messageProcessTime) <= ToleranceSecond
	//		then process the message immediately.
	ToleranceSecond int32 `json:"toleranceSecond"`
}
