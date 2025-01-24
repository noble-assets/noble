- ICA was broken in the v8 Helium upgrade due to IBC migrations that were run,
  which incorrectly updated the host module capabilities to be that of the
  controller module. This fix changes the capability type back to the host
  module. ([\#432](https://github.com/noble-assets/noble/pull/432))