// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract SimpleEvent {
    event MyEvent(uint256 indexed value);

    function emitEvent(uint256 value) public {
        emit MyEvent(value);
    }
}