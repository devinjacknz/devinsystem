package models

import (
	"fmt"
	"log"
	"sync"
)

type ModelFactory struct {
	mu     sync.RWMutex
	models map[string]Model
}

func NewModelFactory() *ModelFactory {
	return &ModelFactory{
		models: make(map[string]Model),
	}
}

func (f *ModelFactory) RegisterModel(modelType string, model Model) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.models[modelType]; exists {
		return fmt.Errorf("model type %s already registered", modelType)
	}

	f.models[modelType] = model
	log.Printf("[SYSTEM] Registered model: %s", modelType)
	return nil
}

func (f *ModelFactory) GetModel(modelType string) (Model, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	model, exists := f.models[modelType]
	if !exists {
		return nil, fmt.Errorf("model type %s not found", modelType)
	}

	if !model.IsAvailable() {
		return nil, fmt.Errorf("model %s is not available", modelType)
	}

	return model, nil
}

func (f *ModelFactory) AvailableModels() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var models []string
	for modelType, model := range f.models {
		if model.IsAvailable() {
			models = append(models, modelType)
		}
	}
	return models
}
