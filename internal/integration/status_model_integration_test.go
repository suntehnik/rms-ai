package integration

import (
	"testing"
)

func TestStatusModelIntegration(t *testing.T) {
	// Skip integration tests if no database is available
	t.Skip("Skipping integration test - requires database setup")

	/*
		// Initialize repositories
		statusModelRepo := repository.NewStatusModelRepository(db)
		statusRepo := repository.NewStatusRepository(db)
		statusTransitionRepo := repository.NewStatusTransitionRepository(db)
		requirementTypeRepo := repository.NewRequirementTypeRepository(db)
		relationshipTypeRepo := repository.NewRelationshipTypeRepository(db)
		requirementRepo := repository.NewRequirementRepository(db)
		requirementRelationRepo := repository.NewRequirementRelationshipRepository(db)

		// Initialize service
		configService := service.NewConfigService(
			requirementTypeRepo,
			relationshipTypeRepo,
			requirementRepo,
			requirementRelationRepo,
			statusModelRepo,
			statusRepo,
			statusTransitionRepo,
		)

		t.Run("Create and retrieve status model", func(t *testing.T) {
			// Create status model
			req := service.CreateStatusModelRequest{
				EntityType:  models.EntityTypeEpic,
				Name:        "Test Epic Workflow",
				Description: stringPtr("Test workflow for epics"),
				IsDefault:   false,
			}

			statusModel, err := configService.CreateStatusModel(req)
			require.NoError(t, err)
			assert.NotNil(t, statusModel)
			assert.Equal(t, models.EntityTypeEpic, statusModel.EntityType)
			assert.Equal(t, "Test Epic Workflow", statusModel.Name)
			assert.False(t, statusModel.IsDefault)

			// Retrieve status model
			retrieved, err := configService.GetStatusModelByID(statusModel.ID)
			require.NoError(t, err)
			assert.Equal(t, statusModel.ID, retrieved.ID)
			assert.Equal(t, statusModel.Name, retrieved.Name)
		})

		t.Run("Create statuses for status model", func(t *testing.T) {
			// Create status model first
			statusModelReq := service.CreateStatusModelRequest{
				EntityType: models.EntityTypeUserStory,
				Name:       "User Story Workflow",
				IsDefault:  true,
			}

			statusModel, err := configService.CreateStatusModel(statusModelReq)
			require.NoError(t, err)

			// Create statuses
			statusReqs := []service.CreateStatusRequest{
				{
					StatusModelID: statusModel.ID,
					Name:          "Backlog",
					Description:   stringPtr("Items in backlog"),
					Color:         stringPtr("#6c757d"),
					IsInitial:     true,
					Order:         1,
				},
				{
					StatusModelID: statusModel.ID,
					Name:          "In Progress",
					Description:   stringPtr("Items being worked on"),
					Color:         stringPtr("#007bff"),
					Order:         2,
				},
				{
					StatusModelID: statusModel.ID,
					Name:          "Done",
					Description:   stringPtr("Completed items"),
					Color:         stringPtr("#28a745"),
					IsFinal:       true,
					Order:         3,
				},
			}

			var createdStatuses []*models.Status
			for _, req := range statusReqs {
				status, err := configService.CreateStatus(req)
				require.NoError(t, err)
				assert.Equal(t, statusModel.ID, status.StatusModelID)
				createdStatuses = append(createdStatuses, status)
			}

			// List statuses for the model
			statuses, err := configService.ListStatusesByModel(statusModel.ID)
			require.NoError(t, err)
			assert.Len(t, statuses, 3)

			// Verify order
			assert.Equal(t, "Backlog", statuses[0].Name)
			assert.True(t, statuses[0].IsInitial)
			assert.Equal(t, "Done", statuses[2].Name)
			assert.True(t, statuses[2].IsFinal)
		})

		t.Run("Create status transitions", func(t *testing.T) {
			// Create status model and statuses
			statusModelReq := service.CreateStatusModelRequest{
				EntityType: models.EntityTypeRequirement,
				Name:       "Requirement Workflow",
				IsDefault:  true,
			}

			statusModel, err := configService.CreateStatusModel(statusModelReq)
			require.NoError(t, err)

			// Create statuses
			draftStatus, err := configService.CreateStatus(service.CreateStatusRequest{
				StatusModelID: statusModel.ID,
				Name:          "Draft",
				IsInitial:     true,
				Order:         1,
			})
			require.NoError(t, err)

			activeStatus, err := configService.CreateStatus(service.CreateStatusRequest{
				StatusModelID: statusModel.ID,
				Name:          "Active",
				Order:         2,
			})
			require.NoError(t, err)

			obsoleteStatus, err := configService.CreateStatus(service.CreateStatusRequest{
				StatusModelID: statusModel.ID,
				Name:          "Obsolete",
				IsFinal:       true,
				Order:         3,
			})
			require.NoError(t, err)

			// Create transitions
			transitions := []service.CreateStatusTransitionRequest{
				{
					StatusModelID: statusModel.ID,
					FromStatusID:  draftStatus.ID,
					ToStatusID:    activeStatus.ID,
					Name:          stringPtr("Activate"),
					Description:   stringPtr("Move from draft to active"),
				},
				{
					StatusModelID: statusModel.ID,
					FromStatusID:  activeStatus.ID,
					ToStatusID:    obsoleteStatus.ID,
					Name:          stringPtr("Obsolete"),
					Description:   stringPtr("Mark as obsolete"),
				},
			}

			for _, req := range transitions {
				transition, err := configService.CreateStatusTransition(req)
				require.NoError(t, err)
				assert.Equal(t, statusModel.ID, transition.StatusModelID)
			}

			// List transitions for the model
			modelTransitions, err := configService.ListStatusTransitionsByModel(statusModel.ID)
			require.NoError(t, err)
			assert.Len(t, modelTransitions, 2)
		})

		t.Run("Validate status transitions", func(t *testing.T) {
			// Create status model with explicit transitions
			statusModelReq := service.CreateStatusModelRequest{
				EntityType: models.EntityTypeEpic,
				Name:       "Restricted Epic Workflow",
				IsDefault:  false,
			}

			statusModel, err := configService.CreateStatusModel(statusModelReq)
			require.NoError(t, err)

			// Create statuses
			backlogStatus, err := configService.CreateStatus(service.CreateStatusRequest{
				StatusModelID: statusModel.ID,
				Name:          "Backlog",
				IsInitial:     true,
			})
			require.NoError(t, err)

			progressStatus, err := configService.CreateStatus(service.CreateStatusRequest{
				StatusModelID: statusModel.ID,
				Name:          "In Progress",
			})
			require.NoError(t, err)

			doneStatus, err := configService.CreateStatus(service.CreateStatusRequest{
				StatusModelID: statusModel.ID,
				Name:          "Done",
				IsFinal:       true,
			})
			require.NoError(t, err)

			// Create only specific transitions (Backlog -> In Progress -> Done)
			_, err = configService.CreateStatusTransition(service.CreateStatusTransitionRequest{
				StatusModelID: statusModel.ID,
				FromStatusID:  backlogStatus.ID,
				ToStatusID:    progressStatus.ID,
			})
			require.NoError(t, err)

			_, err = configService.CreateStatusTransition(service.CreateStatusTransitionRequest{
				StatusModelID: statusModel.ID,
				FromStatusID:  progressStatus.ID,
				ToStatusID:    doneStatus.ID,
			})
			require.NoError(t, err)

			// Test validation - this should work for default models (no explicit transitions)
			err = configService.ValidateStatusTransition(models.EntityTypeUserStory, "Backlog", "Done")
			assert.NoError(t, err) // Should pass because default models allow all transitions
		})

		t.Run("List status models by entity type", func(t *testing.T) {
			// Create multiple status models for different entity types
			epicModel, err := configService.CreateStatusModel(service.CreateStatusModelRequest{
				EntityType: models.EntityTypeEpic,
				Name:       "Epic Model 1",
			})
			require.NoError(t, err)

			userStoryModel, err := configService.CreateStatusModel(service.CreateStatusModelRequest{
				EntityType: models.EntityTypeUserStory,
				Name:       "User Story Model 1",
			})
			require.NoError(t, err)

			// List all status models
			allModels, err := configService.ListStatusModels(service.StatusModelFilters{})
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(allModels), 2)

			// List only epic models
			epicModels, err := configService.ListStatusModels(service.StatusModelFilters{
				EntityType: models.EntityTypeEpic,
			})
			require.NoError(t, err)

			// Should contain at least our created epic model
			found := false
			for _, model := range epicModels {
				if model.ID == epicModel.ID {
					found = true
					break
				}
			}
			assert.True(t, found)

			// Verify user story model is not in epic models list
			for _, model := range epicModels {
				assert.NotEqual(t, userStoryModel.ID, model.ID)
			}
		})

		t.Run("Update status model", func(t *testing.T) {
			// Create status model
			statusModel, err := configService.CreateStatusModel(service.CreateStatusModelRequest{
				EntityType: models.EntityTypeEpic,
				Name:       "Original Name",
				IsDefault:  false,
			})
			require.NoError(t, err)

			// Update status model
			newName := "Updated Name"
			newDescription := "Updated description"
			isDefault := true

			updated, err := configService.UpdateStatusModel(statusModel.ID, service.UpdateStatusModelRequest{
				Name:        &newName,
				Description: &newDescription,
				IsDefault:   &isDefault,
			})
			require.NoError(t, err)
			assert.Equal(t, newName, updated.Name)
			assert.Equal(t, newDescription, *updated.Description)
			assert.True(t, updated.IsDefault)
		})

		t.Run("Delete status model", func(t *testing.T) {
			// Create status model
			statusModel, err := configService.CreateStatusModel(service.CreateStatusModelRequest{
				EntityType: models.EntityTypeEpic,
				Name:       "To Be Deleted",
			})
			require.NoError(t, err)

			// Delete status model
			err = configService.DeleteStatusModel(statusModel.ID, false)
			require.NoError(t, err)

			// Verify it's deleted
			_, err = configService.GetStatusModelByID(statusModel.ID)
			assert.Error(t, err)
			assert.Equal(t, service.ErrStatusModelNotFound, err)
		})
	*/
}
