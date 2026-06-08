package service

import (
	"context"
	"strings"

	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"

	"github.com/google/uuid"
)

type AIService struct {
	jobs       repository.AIRepository
	comfyUIURL string
}

func NewAIService(jobs repository.AIRepository, comfyUIURL string) *AIService {
	return &AIService{
		jobs:       jobs,
		comfyUIURL: strings.TrimSpace(comfyUIURL),
	}
}

func (s *AIService) EnqueueFromMessage(ctx context.Context, message models.PublicMessage) (models.PublicAIJob, bool, error) {
	command, prompt, ok := parseAICommand(message.Content)
	if !ok {
		return models.PublicAIJob{}, false, nil
	}

	job := models.AIJob{
		ID:        uuid.New(),
		ServerID:  message.ServerID,
		ChannelID: message.ChannelID,
		UserID:    message.AuthorID,
		Command:   command,
		Prompt:    prompt,
		Status:    models.AIJobQueued,
		Progress:  0,
	}
	created, err := s.jobs.CreateJob(ctx, job)
	if err != nil {
		return models.PublicAIJob{}, true, err
	}
	return created.Public(), true, nil
}

func parseAICommand(content string) (string, string, bool) {
	content = strings.TrimSpace(content)
	for prefix, command := range map[string]string{
		"/ask ":  models.AICommandAsk,
		"/draw ": models.AICommandDraw,
	} {
		if strings.HasPrefix(content, prefix) {
			prompt := strings.TrimSpace(strings.TrimPrefix(content, prefix))
			return command, prompt, prompt != ""
		}
	}
	return "", "", false
}
