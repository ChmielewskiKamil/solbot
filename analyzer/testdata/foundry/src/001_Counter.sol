// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract Counter {
    uint256 public count;

    // Constructor to initialize the counter
    constructor() {
        count = 0;
    }

    // Function to increment the counter
    function increment() public {
        count += 1;
    }

    // Function to decrement the counter
    function decrement() public {
        require(count > 0, "Counter cannot be negative");
        count -= 1;
    }

    // Function to reset the counter to zero
    function reset() public {
        count = 0;
    }
}
