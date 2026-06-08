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
	servers    repository.ServerRepository
	comfyUIURL string
}

type CreateAIJobInput struct {
	ChannelID uuid.UUID `json:"channel_id"`
	Prompt    string    `json:"prompt"`
}

type AIJobsResponse struct {
	Jobs []models.PublicAIJob `json:"jobs"`
}

func NewAIService(jobs repository.AIRepository, servers repository.ServerRepository, comfyUIURL string) *AIService {
	return &AIService{
		jobs:       jobs,
		servers:    servers,
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

func (s *AIService) CreateJob(ctx context.Context, userID uuid.UUID, command string, input CreateAIJobInput) (models.PublicAIJob, error) {
	input.Prompt = strings.TrimSpace(input.Prompt)
	if fields := validateCreateAIJobInput(command, input); len(fields) > 0 {
		return models.PublicAIJob{}, &ValidationError{Fields: fields}
	}

	channel, err := s.servers.GetChannelForUser(ctx, input.ChannelID, userID)
	if err != nil {
		return models.PublicAIJob{}, mapRepositoryError(err)
	}
	if channel.Type != models.ChannelTypeText {
		return models.PublicAIJob{}, ErrNotFound
	}

	job := models.AIJob{
		ID:        uuid.New(),
		ServerID:  channel.ServerID,
		ChannelID: channel.ID,
		UserID:    userID,
		Command:   command,
		Prompt:    input.Prompt,
		Status:    models.AIJobQueued,
		Progress:  0,
	}
	created, err := s.jobs.CreateJob(ctx, job)
	if err != nil {
		return models.PublicAIJob{}, err
	}
	return created.Public(), nil
}

func (s *AIService) ListJobs(ctx context.Context, userID uuid.UUID, limit int) (AIJobsResponse, error) {
	if limit <= 0 {
		limit = defaultMessageLimit
	}
	if limit > maxMessageLimit {
		limit = maxMessageLimit
	}
	jobs, err := s.jobs.ListJobsForUser(ctx, userID, limit)
	if err != nil {
		return AIJobsResponse{}, mapAIRepositoryError(err)
	}
	return AIJobsResponse{Jobs: publicAIJobs(jobs)}, nil
}

func (s *AIService) GetJob(ctx context.Context, userID, jobID uuid.UUID) (models.PublicAIJob, error) {
	job, err := s.jobs.GetJobForUser(ctx, jobID, userID)
	if err != nil {
		return models.PublicAIJob{}, mapAIRepositoryError(err)
	}
	return job.Public(), nil
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

func mapAIRepositoryError(err error) error {
	if err == nil {
		return nil
	}
	if err == repository.ErrAIJobNotFound {
		return ErrNotFound
	}
	return err
}

func publicAIJobs(jobs []models.AIJob) []models.PublicAIJob {
	public := make([]models.PublicAIJob, 0, len(jobs))
	for _, job := range jobs {
		public = append(public, job.Public())
	}
	return public
}
