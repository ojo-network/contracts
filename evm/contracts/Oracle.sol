// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

contract PriceFeed is Ownable, AccessControl {
    error UnAuthorized();

    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");

    struct Data {
        bytes32 name;
        uint256 value;
        uint256 resolveTime;
        uint256 id;
    }

    struct MedianData{
        bytes32 name;
        uint256[] value;
        uint256 resolveTime;
        uint256 id;
    }

    bool enableWhitelist;
    mapping(address => bool) private whitelist;
    mapping(bytes32 => Data) private prices;
    mapping(bytes32 => Data) private deviations;
    mapping(bytes32 => MedianData) private medians;
    
    bytes32[] public assets;


    event PricePosted(address indexed relayer, uint256 timestamp);
    event DeviationPosted(address indexed relayer, uint256 timestamp);
    event MedianPosted(address indexed relayer, uint256 timestamp);

    event Whitelisted(address user);
    event RemovedFromWhitelist(address user);

    constructor() {
        _setupRole(DEFAULT_ADMIN_ROLE, _msgSender());
        _setupRole(RELAYER_ROLE, _msgSender());
    }

    modifier enableWhitelistOnly{
        if (enableWhitelist){
            if (!whitelist[msg.sender]){
                revert UnAuthorized();
            }
        }
        _;
    }

    function postPrices(Data[] calldata _prices) external onlyRole(RELAYER_ROLE) {
        for(uint256 i=0;i<_prices.length;i++){
            prices[_prices[i].name]= _prices[i];
        }

        emit PricePosted(_msgSender(), block.timestamp);
    }

    function postDeviations(Data[] calldata _deviations) external onlyRole(RELAYER_ROLE) {
        for(uint256 i=0;i<_deviations.length;i++){
            deviations[_deviations[i].name]= _deviations[i];
        }

        emit DeviationPosted(_msgSender(), block.timestamp);
    }


    function postMedians(MedianData[] calldata _medians) external onlyRole(RELAYER_ROLE) {
        for(uint256 i=0;i<_medians.length;i++){
            medians[_medians[i].name]= _medians[i];
        }

        emit MedianPosted(_msgSender(), block.timestamp);
    }

    function getPrice(bytes32 assetName) public view enableWhitelistOnly returns (Data memory) {
        return prices[assetName];
    }

    function getDeviations(bytes32 assetName) public view enableWhitelistOnly returns (Data memory) {
        return deviations[assetName];
    }

    function getMedians(bytes32 assetName) public view enableWhitelistOnly returns (MedianData memory) {
        return medians[assetName];
    }

    function whitelistAddress(address _user) external onlyOwner{
        whitelist[_user]= true;
        emit Whitelisted(_user);
    }

    function removeAddress(address _user) external onlyOwner{
        whitelist[_user]= true;
        emit RemovedFromWhitelist(_user);
    }
    
    function assignRelayerRole(address relayer) public onlyOwner {
        grantRole(RELAYER_ROLE, relayer);
    }

    function revokeRelayerRole(address relayer) public onlyOwner {
        revokeRole(RELAYER_ROLE, relayer);
    }
}