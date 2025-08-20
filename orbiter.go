package noble

import (
	orbiter "github.com/noble-assets/orbiter"
)

func (app *App) RegisterOrbiterControllers() {
	in := orbiter.ComponentsInputs{
		Orbiters:   app.OrbiterKeeper,
		BankKeeper: app.BankKeeper,
		CCTPKeeper: app.CCTPKeeper,
	}

	orbiter.InjectComponents(in)
}
