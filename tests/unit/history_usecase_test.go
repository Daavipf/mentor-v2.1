package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/usecase/history"
	"mentor-api/tests/mocks"
)

func TestHistory_ListAndGetByID_UserRestricted(t *testing.T) {
	qRepo := mocks.NewMockQuestionRepository()
	eRepo := mocks.NewMockExamRepository(qRepo)
	hRepo := mocks.NewMockHistoryRepository()
	historyUC := history.NewHistoryUseCase(hRepo, eRepo)
	ctx := context.Background()

	user1 := "user-1"
	user2 := "user-2"

	scoreVal := 85
	now := time.Now()
	h1 := &entity.History{
		ID:         uuid.New().String(),
		UserID:     user1,
		ExamID:     "exam-1",
		Score:      &scoreVal,
		CreatedAt:  now,
		UpdatedAt:  now,
		FinishedAt: &now,
	}
	require.NoError(t, hRepo.Create(ctx, h1))

	// User 1 lists histories
	user1List, err := historyUC.ListUserHistories(ctx, user1)
	require.NoError(t, err)
	assert.Len(t, user1List, 1)
	assert.Equal(t, h1.ID, user1List[0].ID)

	// User 2 lists histories -> empty
	user2List, err := historyUC.ListUserHistories(ctx, user2)
	require.NoError(t, err)
	assert.Empty(t, user2List)

	// User 1 gets their history
	detail, err := historyUC.GetHistoryByID(ctx, user1, h1.ID)
	require.NoError(t, err)
	assert.Equal(t, h1.ID, detail.ID)
	assert.Equal(t, &scoreVal, detail.Score)

	// User 2 tries to access User 1's history -> MUST BE FORBIDDEN
	_, err = historyUC.GetHistoryByID(ctx, user2, h1.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

