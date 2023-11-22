package keeper_test

import (
	"testing"

<<<<<<< HEAD:x/fiattokenfactory/keeper/params_test.go
	testkeeper "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	testkeeper "github.com/noble-assets/noble/v5/testutil/keeper"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/params_test.go
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.FiatTokenfactoryKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
