package task

import "github.com/EutropicAI/FinalRip/module/queue"

// GetEncodeQueueName 获取队列名称，如果没有则返回默认队列（低优先级）
func GetEncodeQueueName(name *string) string {
	if name == nil {
		return queue.ENCODE_QUEUE_DEFAULT
	}

	switch *name {
	case "default":
		return queue.ENCODE_QUEUE_DEFAULT
	case "priority":
		return queue.ENCODE_QUEUE_PRIORITY
	default:
		return queue.ENCODE_QUEUE_DEFAULT
	}
}
