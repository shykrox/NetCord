package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	AICommandAsk  = "ask"
	AICommandDraw = "draw"
	AIJobQueued   = "queued"
	AIJobRunning  = "running"
	AIJobDone     = "completed"
	AIJobFailed   = "failed"
)

type AIJob struct {
	ID              uuid.UUID
	ServerID        uuid.UUID
	ChannelID       uuid.UUID
	UserID          uuid.UUID
	Command         string
	Prompt          string
	Status          string
	Progress        int
	ResultMessageID *uuid.UUID
	Error           *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     *time.Time
}

type PublicAIJob struct {
	ID              uuid.UUID  `json:"id"`
	ServerID        uuid.UUID  `json:"server_id"`
	ChannelID       uuid.UUID  `json:"channel_id"`
	UserID          uuid.UUID  `json:"user_id"`
	Command         string     `json:"command"`
	Prompt          string     `json:"prompt"`
	Status          string     `json:"status"`
	Progress        int        `json:"progress"`
	ResultMessageID *uuid.UUID `json:"result_message_id"`
	Error           *string    `json:"error"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at"`
}

func (j AIJob) Public() PublicAIJob {
	return PublicAIJob{
		ID:              j.ID,
		ServerID:        j.ServerID,
		ChannelID:       j.ChannelID,
		UserID:          j.UserID,
		Command:         j.Command,
		Prompt:          j.Prompt,
		Status:          j.Status,
		Progress:        j.Progress,
		ResultMessageID: j.ResultMessageID,
		Error:           j.Error,
		CreatedAt:       j.CreatedAt,
		UpdatedAt:       j.UpdatedAt,
		CompletedAt:     j.CompletedAt,
	}
}
