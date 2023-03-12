package db

type ChatMessages struct {
	ID              int64  `json:"id" gorm:"column:id;AUTO_INCREMENT;primary_key"`
	MessageID       string `json:"messageID" gorm:"column:message_id"`
	UserID          int64  `json:"userID" gorm:"column:user_id"`
	AccountNumber   string `json:"accountNumber" gorm:"column:account_number"`
	Type            int    `json:"type" gorm:"column:type"`
	ConversationID  string `json:"conversationID" gorm:"column:conversation_id"`
	ParentMessageID string `json:"parentMessageID" gorm:"column:parent_message_id"`
	Contents        string `json:"contents" gorm:"column:contents"`
	CreatedAt       int64  `json:"createdAt" gorm:"column:created_at"`
	End             bool   `json:"end" gorm:"-"`
	Error           string `json:"error" gorm:"-"`
}

func (chm *ChatMessages) TableName() string {
	return "chat_messages"
}

//,,, implementation of Add() and GetAllByUserId() and GetLastConversationIdAndMessageIdByUserId() by your self
