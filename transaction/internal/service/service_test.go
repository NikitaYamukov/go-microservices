package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"transaction/internal/model"
	"transaction/internal/service/mocks"
)

func TestTransactionService_Deposit(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint64
		amount         float64
		setupMocks     func(*mocks.MockRepository, *mocks.MockKafkaPublisher)
		expectedError  bool
		expectedResult model.TransactionDetails
	}{
		{
			name:   "successful deposit",
			userID: 1,
			amount: 100.50,
			setupMocks: func(mockRepo *mocks.MockRepository, mockKafka *mocks.MockKafkaPublisher) {
				expectedParams := model.DepositParams{
					UserID: 1,
					Amount: 100.50,
				}

				expectedResult := model.TransactionDetails{
					Transaction: model.Transaction{
						ID:        1,
						UserID:    1,
						Amount:    10050,
						Status:    model.TransactionStatusPending,
						Type:      model.TransactionTypeDeposit,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Entries: []model.TransactionEntry{},
				}

				mockRepo.EXPECT().
					Deposit(gomock.Any(), expectedParams).
					Return(expectedResult, nil)

				mockKafka.EXPECT().
					Publish(gomock.Any(), "transaction_data", "1", gomock.Any()).
					Return(nil)
			},
			expectedError: false,
			expectedResult: model.TransactionDetails{
				Transaction: model.Transaction{
					ID:        1,
					UserID:    1,
					Amount:    10050,
					Status:    model.TransactionStatusPending,
					Type:      model.TransactionTypeDeposit,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Entries: []model.TransactionEntry{},
			},
		},
		{
			name:   "repository error",
			userID: 1,
			amount: 100.50,
			setupMocks: func(mockRepo *mocks.MockRepository, mockKafka *mocks.MockKafkaPublisher) {
				expectedParams := model.DepositParams{
					UserID: 1,
					Amount: 100.50,
				}

				mockRepo.EXPECT().
					Deposit(gomock.Any(), expectedParams).
					Return(model.TransactionDetails{}, errors.New("database error"))
			},
			expectedError:  true,
			expectedResult: model.TransactionDetails{},
		},
		{
			name:   "kafka publish error",
			userID: 1,
			amount: 100.50,
			setupMocks: func(mockRepo *mocks.MockRepository, mockKafka *mocks.MockKafkaPublisher) {
				expectedParams := model.DepositParams{
					UserID: 1,
					Amount: 100.50,
				}

				expectedResult := model.TransactionDetails{
					Transaction: model.Transaction{
						ID:        1,
						UserID:    1,
						Amount:    10050,
						Status:    model.TransactionStatusPending,
						Type:      model.TransactionTypeDeposit,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Entries: []model.TransactionEntry{},
				}

				mockRepo.EXPECT().
					Deposit(gomock.Any(), expectedParams).
					Return(expectedResult, nil)

				mockKafka.EXPECT().
					Publish(gomock.Any(), "transaction_data", "1", gomock.Any()).
					Return(errors.New("kafka error"))

				mockRepo.EXPECT().
					UpdateTransactionStatus(gomock.Any(), uint64(1), model.TransactionStatusFailed).
					Return(nil)
			},
			expectedError:  true,
			expectedResult: model.TransactionDetails{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockAccountService := mocks.NewMockAccountService(ctrl)
			mockKafka := mocks.NewMockKafkaPublisher(ctrl)
			logger := zerolog.Nop()

			tt.setupMocks(mockRepo, mockKafka)

			service := New(mockRepo, mockAccountService, mockKafka, &logger)

			result, err := service.Deposit(context.Background(), tt.userID, tt.amount)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Transaction.ID, result.Transaction.ID)
				assert.Equal(t, tt.expectedResult.Transaction.UserID, result.Transaction.UserID)
				assert.Equal(t, tt.expectedResult.Transaction.Amount, result.Transaction.Amount)
				assert.Equal(t, tt.expectedResult.Transaction.Status, result.Transaction.Status)
				assert.Equal(t, tt.expectedResult.Transaction.Type, result.Transaction.Type)
			}
		})
	}
}

func TestTransactionService_Withdraw(t *testing.T) {
	tests := []struct {
		name           string
		accountID      uint64
		amount         float64
		setupMocks     func(*mocks.MockRepository, *mocks.MockKafkaPublisher)
		expectedError  bool
		expectedResult model.TransactionDetails
	}{
		{
			name:      "successful withdraw",
			accountID: 1,
			amount:    50.25,
			setupMocks: func(mockRepo *mocks.MockRepository, mockKafka *mocks.MockKafkaPublisher) {
				expectedParams := model.WithdrawParams{
					AccountID: 1,
					Amount:    50.25,
				}

				expectedResult := model.TransactionDetails{
					Transaction: model.Transaction{
						ID:        2,
						UserID:    1,
						Amount:    5025,
						Status:    model.TransactionStatusPending,
						Type:      model.TransactionTypeWithdraw,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Entries: []model.TransactionEntry{},
				}

				mockRepo.EXPECT().
					Withdraw(gomock.Any(), expectedParams).
					Return(expectedResult, nil)

				mockKafka.EXPECT().
					Publish(gomock.Any(), "transaction_data", "1", gomock.Any()).
					Return(nil)
			},
			expectedError: false,
			expectedResult: model.TransactionDetails{
				Transaction: model.Transaction{
					ID:        2,
					UserID:    1,
					Amount:    5025,
					Status:    model.TransactionStatusPending,
					Type:      model.TransactionTypeWithdraw,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Entries: []model.TransactionEntry{},
			},
		},
		{
			name:      "repository error",
			accountID: 1,
			amount:    50.25,
			setupMocks: func(mockRepo *mocks.MockRepository, mockKafka *mocks.MockKafkaPublisher) {
				expectedParams := model.WithdrawParams{
					AccountID: 1,
					Amount:    50.25,
				}

				mockRepo.EXPECT().
					Withdraw(gomock.Any(), expectedParams).
					Return(model.TransactionDetails{}, errors.New("database error"))
			},
			expectedError:  true,
			expectedResult: model.TransactionDetails{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockAccountService := mocks.NewMockAccountService(ctrl)
			mockKafka := mocks.NewMockKafkaPublisher(ctrl)
			logger := zerolog.Nop()

			tt.setupMocks(mockRepo, mockKafka)

			service := New(mockRepo, mockAccountService, mockKafka, &logger)

			result, err := service.Withdraw(context.Background(), tt.accountID, tt.amount)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Transaction.ID, result.Transaction.ID)
				assert.Equal(t, tt.expectedResult.Transaction.UserID, result.Transaction.UserID)
				assert.Equal(t, tt.expectedResult.Transaction.Amount, result.Transaction.Amount)
				assert.Equal(t, tt.expectedResult.Transaction.Status, result.Transaction.Status)
				assert.Equal(t, tt.expectedResult.Transaction.Type, result.Transaction.Type)
			}
		})
	}
}

func TestTransactionService_Transfer(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint64
		recipient      uint64
		amount         float64
		setupMocks     func(*mocks.MockRepository, *mocks.MockKafkaPublisher)
		expectedError  bool
		expectedResult model.TransactionDetails
	}{
		{
			name:      "successful transfer",
			userID:    1,
			recipient: 2,
			amount:    75.00,
			setupMocks: func(mockRepo *mocks.MockRepository, mockKafka *mocks.MockKafkaPublisher) {
				expectedParams := model.TransferParams{
					UserID:    1,
					Recipient: 2,
					Amount:    75.00,
				}

				expectedResult := model.TransactionDetails{
					Transaction: model.Transaction{
						ID:        3,
						UserID:    1,
						Amount:    7500,
						Status:    model.TransactionStatusPending,
						Type:      model.TransactionTypeTransfer,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Entries: []model.TransactionEntry{},
				}

				mockRepo.EXPECT().
					Transfer(gomock.Any(), expectedParams).
					Return(expectedResult, nil)

				mockKafka.EXPECT().
					Publish(gomock.Any(), "transaction_data", "1", gomock.Any()).
					Return(nil)
			},
			expectedError: false,
			expectedResult: model.TransactionDetails{
				Transaction: model.Transaction{
					ID:        3,
					UserID:    1,
					Amount:    7500,
					Status:    model.TransactionStatusPending,
					Type:      model.TransactionTypeTransfer,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Entries: []model.TransactionEntry{},
			},
		},
		{
			name:      "repository error",
			userID:    1,
			recipient: 2,
			amount:    75.00,
			setupMocks: func(mockRepo *mocks.MockRepository, mockKafka *mocks.MockKafkaPublisher) {
				expectedParams := model.TransferParams{
					UserID:    1,
					Recipient: 2,
					Amount:    75.00,
				}

				mockRepo.EXPECT().
					Transfer(gomock.Any(), expectedParams).
					Return(model.TransactionDetails{}, errors.New("database error"))
			},
			expectedError:  true,
			expectedResult: model.TransactionDetails{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockAccountService := mocks.NewMockAccountService(ctrl)
			mockKafka := mocks.NewMockKafkaPublisher(ctrl)
			logger := zerolog.Nop()

			tt.setupMocks(mockRepo, mockKafka)

			service := New(mockRepo, mockAccountService, mockKafka, &logger)

			result, err := service.Transfer(context.Background(), tt.userID, tt.recipient, tt.amount)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Transaction.ID, result.Transaction.ID)
				assert.Equal(t, tt.expectedResult.Transaction.UserID, result.Transaction.UserID)
				assert.Equal(t, tt.expectedResult.Transaction.Amount, result.Transaction.Amount)
				assert.Equal(t, tt.expectedResult.Transaction.Status, result.Transaction.Status)
				assert.Equal(t, tt.expectedResult.Transaction.Type, result.Transaction.Type)
			}
		})
	}
}

func TestTransactionService_HandleAccountResponse(t *testing.T) {
	tests := []struct {
		name          string
		response      AccountResponse
		setupMocks    func(*mocks.MockRepository)
		expectedError bool
	}{
		{
			name: "successful response - transaction completed",
			response: AccountResponse{
				RequestType: "deposit",
				UserID:      1,
				OperationID: 1,
				Result:      true,
			},
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.EXPECT().
					UpdateTransactionStatus(gomock.Any(), uint64(1),
						model.TransactionStatusCompleted).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name: "failed response - transaction failed",
			response: AccountResponse{
				RequestType: "withdraw",
				UserID:      1,
				OperationID: 2,
				Result:      false,
			},
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.EXPECT().
					UpdateTransactionStatus(gomock.Any(), uint64(2), model.TransactionStatusFailed).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name: "repository error",
			response: AccountResponse{
				RequestType: "deposit",
				UserID:      1,
				OperationID: 1,
				Result:      true,
			},
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.EXPECT().
					UpdateTransactionStatus(gomock.Any(), uint64(1),
						model.TransactionStatusCompleted).
					Return(errors.New("database error"))
			},
			expectedError: true,
		},
		{
			name: "invalid response - missing request type",
			response: AccountResponse{
				RequestType: "",
				UserID:      1,
				OperationID: 1,
				Result:      true,
			},
			setupMocks: func(mockRepo *mocks.MockRepository) {
				// No repository calls expected for invalid response
			},
			expectedError: true,
		},
		{
			name: "invalid response - missing user ID",
			response: AccountResponse{
				RequestType: "deposit",
				UserID:      0,
				OperationID: 1,
				Result:      true,
			},
			setupMocks: func(mockRepo *mocks.MockRepository) {
				// No repository calls expected for invalid response
			},
			expectedError: true,
		},
		{
			name: "invalid response - missing operation ID",
			response: AccountResponse{
				RequestType: "deposit",
				UserID:      1,
				OperationID: 0,
				Result:      true,
			},
			setupMocks: func(mockRepo *mocks.MockRepository) {
				// No repository calls expected for invalid response
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockAccountService := mocks.NewMockAccountService(ctrl)
			mockKafka := mocks.NewMockKafkaPublisher(ctrl)
			logger := zerolog.Nop()

			tt.setupMocks(mockRepo)

			service := New(mockRepo, mockAccountService, mockKafka, &logger)

			// Convert response to JSON bytes
			responseBytes, err := json.Marshal(tt.response)
			assert.NoError(t, err)

			err = service.HandleAccountResponse(context.Background(), "test_topic", "test_key",
				responseBytes)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransactionService_GetTransactionsWithDetails(t *testing.T) {
	tests := []struct {
		name          string
		params        model.GetTransactionsParams
		setupMocks    func(*mocks.MockRepository)
		expectedError bool
		expectedCount int
	}{
		{
			name: "successful get transactions",
			params: model.GetTransactionsParams{
				UserID: func() *uint64 { id := uint64(1); return &id }(),
				Limit:  10,
				Offset: 0,
			},
			setupMocks: func(mockRepo *mocks.MockRepository) {
				transactions := []model.Transaction{
					{
						ID:        1,
						UserID:    1,
						Amount:    10000,
						Status:    model.TransactionStatusCompleted,
						Type:      model.TransactionTypeDeposit,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					{
						ID:        2,
						UserID:    1,
						Amount:    5000,
						Status:    model.TransactionStatusCompleted,
						Type:      model.TransactionTypeWithdraw,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}

				mockRepo.EXPECT().
					GetTransactions(gomock.Any(), gomock.Any()).
					Return(transactions, nil)

				// Mock GetTransactionDetails for each transaction
				for _, tx := range transactions {
					details := model.TransactionDetails{
						Transaction: tx,
						Entries:     []model.TransactionEntry{},
					}

					mockRepo.EXPECT().
						GetTransactionDetails(gomock.Any(), tx.ID).
						Return(details, nil)
				}
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name: "repository error on get transactions",
			params: model.GetTransactionsParams{
				UserID: func() *uint64 { id := uint64(1); return &id }(),
				Limit:  10,
				Offset: 0,
			},
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.EXPECT().
					GetTransactions(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedError: true,
			expectedCount: 0,
		},
		{
			name: "error on get transaction details",
			params: model.GetTransactionsParams{
				UserID: func() *uint64 { id := uint64(1); return &id }(),
				Limit:  10,
				Offset: 0,
			},
			setupMocks: func(mockRepo *mocks.MockRepository) {
				transactions := []model.Transaction{
					{
						ID:        1,
						UserID:    1,
						Amount:    10000,
						Status:    model.TransactionStatusCompleted,
						Type:      model.TransactionTypeDeposit,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}

				mockRepo.EXPECT().
					GetTransactions(gomock.Any(), gomock.Any()).
					Return(transactions, nil)

				mockRepo.EXPECT().
					GetTransactionDetails(gomock.Any(), uint64(1)).
					Return(model.TransactionDetails{}, errors.New("details error"))
			},
			expectedError: false, // Method continues despite individual detail errors
			expectedCount: 0,     // But returns empty slice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockAccountService := mocks.NewMockAccountService(ctrl)
			mockKafka := mocks.NewMockKafkaPublisher(ctrl)
			logger := zerolog.Nop()

			tt.setupMocks(mockRepo)

			service := New(mockRepo, mockAccountService, mockKafka, &logger)

			result, err := service.GetTransactionsWithDetails(context.Background(), tt.params)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}
