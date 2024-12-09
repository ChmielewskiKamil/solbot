| Detector ID | Description | Implemented |
| --- | --- | :---: |
| `screamingsnakeconst` | `constant` and `immutable` variables should be declared with a `SCREAMING_SNAKE_CASE`. | ✅ |
| `nonpausable` | Contract is not pausable if the internal `_pause` and `_unpause` functions are not exposed | |
| `disableinitializers` | Initializers on implementation contracts should be disabled | |
| `interfacemismatch` | Function signature in the interface is different from the implementation | |
| `zeroaddresseth` | `address(0)` should not be used to represent Ether | |
| `functionorder` | Order of functions should follow the Solidity style guide | |
| `privatefuncunderscore` | Private and internal functions should be prefixed with an underscore | |
| `privatevarunderscore` | Private and internal state variables should be prefixed with an underscore | |
| `renounceownership` | If the `renounceOwnership(...)` is not overriden, ownership can be lost by accident | | 
| `unusedpayable` | Function is marked as `payable` but does not use the `msg.value` inside the function's body | |
| `memorytocalldata` | If function arguments are not modified in the function, they should be declared as `calldata` | |
| `rtloverride` | The `U+202E` character should not be present in the code | |
| `constantvars` | Variables that never change should be declared as `constant` | |
| `immutablevars` | Variables that are assigned once during construction should be declared as `immutable` | |
| `publicexternalfunc` | A `public` function that is not called internally should be declared as `external` | |
| `unusedimport` | Unused imports should be removed | |
| `unusedlocalvar` | Unused local variables should be removed | |
| `unusedstatevar` | Unused state variables should be removed | |
| `unusedreturn` | Unused named returns should be removed | |
| `unusedstruct` | Unused structs should be removed | |
| `unusedmodifier` | Unused modifiers should be removed | |
| `unusedevent` | Unused events should be removed | |
| `unusedenum` | Unused enums should be removed | |
| `unusedfunction` | Unused `internal` and `private` functions should be removed | |
| `unusedparams` | Unused function parameters should be removed | |
| `redefinedconst` | Redefined `constant` and `immutable` variables should be grouped in a single file | |
| `couldbepure` | Functions that do not read or modify state should be declared as `pure` | |
| `unnecessarysetroleadmin` | When using OZ's `AccessControl` there is no need to set `DEFAULT_ADMIN_ROLE` as admin for other roles | |
| `grantrolezeroaddress` | The grant role functions from OZ's `AccessControl` don't check zero address | |
