package processor

import (
	"context"
	"errors"
	"testing"

	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/notifier"
	"github.com/PxyUp/fitter/pkg/parser"
)

type failingEngine struct{}

func (failingEngine) Get(_ context.Context, _ *config.Model, _ builder.Interfacable, _ *uint32, _ builder.Interfacable) (*parser.ParseResult, error) {
	return nil, errors.New("the site is gone")
}

// A failing scrape with an array config and a by-item notifier used to
// evaluate result.IsEmpty() on the nil result the engine returned — and the
// run that most needed its failure notified was the one that crashed the
// whole process instead.
func TestAFailingArrayScrapeNotifiesInsteadOfCrashing(t *testing.T) {
	cfg := &config.NotifierConfig{SendArrayByItem: true, Console: &config.ConsoleConfig{}}
	p := New("failing", failingEngine{}, &config.Model{IsArray: true},
		notifier.NewConsole("failing", cfg.Console), cfg).WithLogger(logger.Null)

	_, err := p.Process(context.Background(), nil)
	if err == nil {
		t.Fatal("a failed scrape reported success")
	}
}

// New is exported with notifier and notifierCfg as independent parameters,
// so the pairing CreateProcessor happens to guarantee is not a contract.
// A notifier without its config must degrade, not dereference.
func TestANotifierWithoutItsConfigDoesNotPanic(t *testing.T) {
	p := New("bare", failingEngine{}, &config.Model{IsArray: true},
		notifier.NewConsole("bare", &config.ConsoleConfig{}), nil).WithLogger(logger.Null)

	if _, err := p.Process(context.Background(), nil); err == nil {
		t.Fatal("a failed scrape reported success")
	}
}
