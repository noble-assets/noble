package e2e_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/noble-assets/noble/e2e"
	"github.com/stretchr/testify/require"
)

func TestRestrictedModules(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	nw := e2e.NobleSpinUp(t, ctx, false)
	noble := nw.Chain.GetNode()

	restrictedModules := []string{"circuit", "gov", "group"}

	for _, module := range restrictedModules {
		require.False(t, noble.HasCommand(ctx, "query", module), fmt.Sprintf("%s is a restricted module", module))
	}
}
