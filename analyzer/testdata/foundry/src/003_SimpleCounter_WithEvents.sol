// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

event OutsideOfContract(uint256 number);

contract Counter {
    event InsideOfContract(uint256 number);
    uint256 public count;

    // Comment?
    constructor() {
        count = 100;
    }

    function increment() public {
        count += 1;
        emit InsideOfContract(count);
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
