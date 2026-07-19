package service

import (
	"context"
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/config"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/events"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository/postgres"
)

type LoanService struct {
	repos     *postgres.Repos
	cfg       config.Config
	publisher events.Publisher
}

func NewLoanService(repos *postgres.Repos, cfg config.Config, publisher events.Publisher) *LoanService {
	if publisher == nil {
		publisher = events.NewNoopPublisher()
	}
	return &LoanService{
		repos:     repos,
		cfg:       cfg,
		publisher: publisher,
	}
}

func (s *LoanService) Issue(ctx context.Context, readerID, copyID int64) (*domain.Loan, error) {
	if readerID <= 0 || copyID <= 0 {
		return nil, domain.ErrInvalidArgument
	}

	var loan *domain.Loan
	err := s.repos.WithTx(ctx, func(txCtx context.Context) error {
		reader, err := s.repos.Readers.GetByID(txCtx, readerID)
		if err != nil {
			return err
		}
		if reader.Status != domain.ReaderActive {
			return domain.ErrReaderBlocked
		}

		activeCount, err := s.repos.Loans.CountActiveByReader(txCtx, readerID)
		if err != nil {
			return err
		}
		if activeCount >= s.cfg.MaxLoansPerReader {
			return domain.ErrLoanLimitExceeded
		}

		copyItem, err := s.repos.Copies.GetByIDForUpdate(txCtx, copyID)
		if err != nil {
			return err
		}
		if copyItem.Status != domain.CopyAvailable {
			return domain.ErrCopyUnavailable
		}

		created := domain.NewLoan(copyID, readerID, s.cfg.LoanDuration)
		id, err := s.repos.Loans.Create(txCtx, created)
		if err != nil {
			return err
		}
		if err := s.repos.Copies.UpdateStatus(txCtx, copyID, domain.CopyOnLoan); err != nil {
			return err
		}

		loan, err = s.repos.Loans.GetByID(txCtx, id)
		return err
	})
	if err != nil {
		return nil, err
	}

	_ = s.publisher.Publish(ctx, events.Event{
		Type:    events.EventLoanIssued,
		Payload: loan,
	})
	return loan, nil
}

func (s *LoanService) Return(ctx context.Context, loanID int64) (*domain.Loan, error) {
	if loanID <= 0 {
		return nil, domain.ErrInvalidArgument
	}

	var loan *domain.Loan
	err := s.repos.WithTx(ctx, func(txCtx context.Context) error {
		current, err := s.repos.Loans.GetByID(txCtx, loanID)
		if err != nil {
			return err
		}
		if current.Status != domain.LoanActive && current.Status != domain.LoanOverdue {
			return domain.ErrLoanNotActive
		}

		if err := s.repos.Loans.MarkReturned(txCtx, loanID); err != nil {
			return err
		}
		if err := s.repos.Copies.UpdateStatus(txCtx, current.CopyID, domain.CopyAvailable); err != nil {
			return err
		}

		loan, err = s.repos.Loans.GetByID(txCtx, loanID)
		return err
	})
	if err != nil {
		return nil, err
	}

	_ = s.publisher.Publish(ctx, events.Event{
		Type:    events.EventLoanReturned,
		Payload: loan,
	})
	return loan, nil
}

func (s *LoanService) ListByReader(ctx context.Context, readerID int64) ([]domain.Loan, error) {
	if readerID <= 0 {
		return nil, domain.ErrInvalidArgument
	}
	return s.repos.Loans.ListByReader(ctx, readerID)
}

func (s *LoanService) GetByID(ctx context.Context, id int64) (*domain.Loan, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidArgument
	}
	loan, err := s.repos.Loans.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if loan.Status == domain.LoanActive && time.Now().UTC().After(loan.DueAt) {
		loan.Status = domain.LoanOverdue
	}
	return loan, nil
}
