package domain

import "time"

type Status string

const (
	StatusPending Status = "PENDING"
	StatusSending Status = "SENDING"
	StatusSent    Status = "SENT"
	StatusFailed  Status = "FAILED"
)

type Channel string

const ChannelSMS Channel = "SMS"


type Priority int

const (
	PriorityLow Priority = 0
	PriorityNormal Priority = 1
	PriorityHigh Priority = 2
)

type Message struct {
	ID             string    // UUID assigned at creation time
	IdempotencyKey string    // caller-supplied dedup key; unique index in store
	Recipient      string    // phone number in E.164 format, e.g. +989121234567
	Body           string    // SMS text, max 1600 chars (10 segments)
	Channel        Channel   // always ChannelSMS for now
	Priority       Priority  // influences worker dispatch order
	Status         Status    // current lifecycle state
	Attempts       int       // how many send attempts have been made
	MaxAttempts    int       // retry ceiling; default set by NotificationService
	ProviderMsgID  string    // ID returned by the SMS gateway after acceptance
	FailureReason  string    // last provider error message, populated on FAILED
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
