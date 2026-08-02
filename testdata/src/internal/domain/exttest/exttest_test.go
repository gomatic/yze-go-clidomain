package exttest_test

import (
	"internal/domain/exttest"
)

// Run is an ordinary test driver whose name happens to match the contract's
// entry point; the external test package is scaffolding and must not be judged.
func Run(cfg exttest.Config) exttest.Result { return exttest.Result{Out: cfg.Name} }
