package rag

// applyLoadModePolicy enforces the pre-load policy for a given DocumentLoadMode.
// Returns true if document loading should proceed, false if it should be skipped.
// Logs errors/debug messages via agent.log and resets the store when mode is Overwrite.
func applyLoadModePolicy(agent *BaseAgent, loadMode DocumentLoadMode, hasExistingData bool) bool {
	switch loadMode {
	case DocumentLoadModeError:
		if hasExistingData {
			agent.log.Error("Store already contains data and DocumentLoadModeError is set")
			return false
		}

	case DocumentLoadModeSkip:
		if hasExistingData {
			agent.log.Debug("Store already contains data, skipping document loading (DocumentLoadModeSkip)")
			return false
		}

	case DocumentLoadModeSkipDuplicates:
		if hasExistingData {
			agent.log.Debug("Checking documents individually for duplicates (DocumentLoadModeSkipDuplicates)")
		}

	case DocumentLoadModeOverwrite:
		return handleOverwriteMode(agent, hasExistingData)

	case DocumentLoadModeMerge:
		if hasExistingData {
			agent.log.Debug("Merging documents with existing data (DocumentLoadModeMerge)")
		}
	}
	return true
}

// handleOverwriteMode clears the store when DocumentLoadModeOverwrite is active and the store
// is non-empty. Returns true if loading should proceed, false if reset failed or is unsupported.
func handleOverwriteMode(agent *BaseAgent, hasExistingData bool) bool {
	if !hasExistingData {
		return true
	}
	resettable, ok := agent.store.(interface{ ResetMemory() error })
	if !ok {
		agent.log.Warn("Store does not support reset, cannot overwrite existing data")
		return false
	}
	if err := resettable.ResetMemory(); err != nil {
		agent.log.Error("Failed to reset store: %v", err)
		return false
	}
	agent.log.Debug("Store cleared (DocumentLoadModeOverwrite)")
	return true
}

// loadDocumentsIntoStore iterates over documents, optionally skipping duplicates,
// and saves each one via GenerateThenSaveEmbeddingVector.
// Returns the number of documents added and the number skipped.
func loadDocumentsIntoStore(agent *BaseAgent, documents []string, loadMode DocumentLoadMode) (added, skipped int) {
	for idx, doc := range documents {
		if loadMode == DocumentLoadModeSkipDuplicates {
			exists, err := documentExistsInStore(agent.store, doc)
			if err != nil {
				agent.log.Error("Failed to check document existence for document %d: %v", idx, err)
				continue
			}
			if exists {
				agent.log.Debug("Document %d already exists, skipping (duplicate)", idx)
				skipped++
				continue
			}
		}

		if err := agent.GenerateThenSaveEmbeddingVector(doc); err != nil {
			agent.log.Error("Failed to save embedding for document %d: %v", idx, err)
		} else {
			agent.log.Debug("Successfully saved embedding for document %d", idx)
			added++
		}
	}
	return added, skipped
}
