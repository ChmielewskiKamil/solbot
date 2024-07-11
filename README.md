| Detector ID | Description | Current blockers | Implemented |
| --- | --- | --- | :---: |
| `screamingsnakeconst` | `constant` and `immutable` variables should be declared with a `SCREAMING_SNAKE_CASE`. | | ✅ |
| `nonpausable` | Contract is not pausable if the internal `_pause` and `_unpause` functions are not exposed | Inheritance parsing; Dependency resolution; Expression parsing | |
| `unusedlocalvar` | Unused local variables should be removed | Expression parsing | |
| `unusedstatevar` | Unused state variables should be removed | Inheritance parsing; Dependency resolution; Expression parsing | |
| `disableinitializers` | Initializers on implementation contracts should be disabled | Expression parsing; Function signature parsing | |
| `interfacemismatch` | Function signature in the interface is different from the implementation | Inheritance parsing; Dependency resolution; Function signature parsing | |
| `zeroaddresseth` | `address(0)` should not be used to represent Ether | Expression parsing | |
| `functionorder` | Order of functions should follow the Solidity style guide | Function signature parsing | |
| `privatefuncunderscore` | Private and internal functions should be prefixed with an underscore | Function signature parsing | |
| `privatevarunderscore` | Private and internal state variables should be prefixed with an underscore | Inheritance parsing; Dependency resolution; Expression parsing | |
| `renounceownership` | If the `renounceOwnership(...)` is not overriden, ownership can be lost by accident | Inheritance parsing; Dependency resolution; Function signature parsing | | 
| `unusedpayable` | Function is marked as `payable` but does not use the `msg.value` inside the function's body | Function signature parsing; Expression parsing | |
| `memorytocalldata` | If function arguments are not modified in the function, they should be declared as `calldata` | Function signature parsing; Expression parsing | |
