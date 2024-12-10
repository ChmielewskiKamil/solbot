// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract Counter {
    uint256 public count;

    // Comment?
    constructor() {
        count = 0;
    }

    function increment() public {
        count += 1;
    }

    function decrement() public {
        count -= 1;
    }

    // Function to reset the counter to zero
    function reset() public {
        if (count != 0) {
            count = 0;
        }
    }
}
