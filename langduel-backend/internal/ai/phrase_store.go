package ai

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"langduel/internal/storage"
)

type PhraseStore struct {
	repo      *storage.DuelRepo
	generator *Generator
}

func NewPhraseStore(repo *storage.DuelRepo) *PhraseStore {
	return &PhraseStore{
		repo:      repo,
		generator: NewGenerator(),
	}
}

func (s *PhraseStore) GenerateAndStore(ctx context.Context, duelID, roomCode, topic, difficulty, langFrom, langTo string) error {
	phrases, err := s.generator.GeneratePhrases(ctx, topic, difficulty, langFrom, langTo, 25)
	if err != nil {
		return fmt.Errorf("failed to generate phrases: %w", err)
	}

	for _, p := range phrases {
		answers := make([]string, len(p.Answers))
		copy(answers, p.Answers)
		rand.Shuffle(len(answers), func(i, j int) {
			answers[i], answers[j] = answers[j], answers[i]
		})

		err := s.repo.SaveAIPhrase(ctx, duelID, roomCode, p.Prompt, answers, topic, difficulty, langFrom, langTo)
		if err != nil {
			return fmt.Errorf("failed to save phrase: %w", err)
		}
	}

	return nil
}

func (s *PhraseStore) GetAIPhrases(ctx context.Context, duelID, roomCode string) ([]storage.AIPhrase, error) {
	return s.repo.GetAIPhrases(ctx, duelID, roomCode)
}

func (s *PhraseStore) UseFallback(topic, difficulty string) []storage.Phrase {
	return storage.GetPhrases(topic, difficulty)
}

func (s *PhraseStore) InitFallback() {
	rand.Seed(time.Now().UnixNano())
}
