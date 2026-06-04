package account

import (
	"context"
	"fmt"
	"strconv"

	accountpb "github.com/NikitaYamukov/contracts/account/go"
)

type Service struct {
	client accountpb.AccountClient
}

func New(client accountpb.AccountClient) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) GetBalance(ctx context.Context, userID uint64) (int64, error) {
	res, err := s.client.GetBalance(ctx, &accountpb.GetBalanceRequest{
		UserId: userID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get account balance: %w", err)
	}

	return res.Balance, nil
}

func (s *Service) Deposit(ctx context.Context, userID uint64, amount int64, operationID uint64) (int64, error) {
	res, err := s.client.Deposit(ctx, &accountpb.DepositRequest{
		UserId:      userID,
		Amount:      amount,
		OperationId: strconv.FormatUint(operationID, 10),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to deposit account: %w", err)
	}

	return res.Balance, nil
}

func (s *Service) Withdraw(ctx context.Context, userID uint64, amount int64, operationID uint64) (int64, error) {
	res, err := s.client.Withdraw(ctx, &accountpb.WithdrawRequest{
		UserId:      userID,
		Amount:      amount,
		OperationId: strconv.FormatUint(operationID, 10),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to withdraw account: %w", err)
	}

	return res.Balance, nil
}

func (s *Service) Transfer(ctx context.Context, userID, recipient uint64, amount int64, operationID uint64) (string, error) {
	res, err := s.client.Transfer(ctx, &accountpb.TransferRequest{
		UserId:      userID,
		RecipientId: recipient,
		Amount:      amount,
		OperationId: strconv.FormatUint(operationID, 10),
	})
	if err != nil {
		return "", fmt.Errorf("failed to transfer account: %w", err)
	}

	return res.Status, nil
}
