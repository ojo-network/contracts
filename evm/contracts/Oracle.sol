// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

contract PriceFeed is Ownable, AccessControl {
    error UnAuthorized();
    error MedianDisabled();

    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");
    bytes32 public constant USD = bytes32("USD");
    uint256 constant USD_PRICE= 10**9;

    struct Data {
        bytes32 name;
        uint256 value;
        uint256 resolveTime;
        uint256 id;
    }

    struct MedianData{
        bytes32 name;
        uint256[] values;
        uint256 resolveTime;
        uint256 id;
    }

    struct Price{
        uint256 price;
        uint256 baseResolveTime;
        uint256 quoteResolveTime;
    }

    bool whitelistEnabled;
    bool medianDisabled;
    mapping(address => bool) private whitelist;
    mapping(bytes32 => Data) private prices;
    mapping(bytes32 => Data) private deviations;
    mapping(bytes32 => MedianData) private medians;

    event PricePosted(address indexed relayer, uint256 indexed timestamp);
    event DeviationPosted(address indexed relayer, uint256 indexed timestamp);
    event MedianPosted(address indexed relayer, uint256 indexed timestamp);

    event Whitelisted(address indexed user);
    event EnableWhitelist(bool indexed status);
    event RemovedFromWhitelist(address indexed user);

    constructor() {
        _setupRole(DEFAULT_ADMIN_ROLE, _msgSender());
        _setupRole(RELAYER_ROLE, _msgSender());
    }

    modifier whitelistCheck{
        if (whitelistEnabled){
            if (!whitelist[_msgSender()]){
                revert UnAuthorized();
            }
        }
        _;
    }

    modifier medianCheck{
        if (medianDisabled){
            revert MedianDisabled();
        }
        _;
    }

    function setWhitelistStatus(bool _status) external onlyOwner{
        whitelistEnabled=_status;

        emit EnableWhitelist(_status);
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

    function getPriceData(bytes32 _assetName) external view whitelistCheck returns (Data memory) {
        return _getPriceData(_assetName);
    }


    function _getPriceData(bytes32 _assetName) internal view returns (Data memory){
        if (_assetName==USD){
            return  Data({
                name:USD, 
                value: USD_PRICE,
                resolveTime: 2**256-1,
                id:0
        });
        }

        return prices[_assetName];
    }

    function getDeviationData(bytes32 _assetName) public view whitelistCheck returns (Data memory) {
        if (_assetName==USD){
            return Data ({
                name:USD, 
                value:0,
                resolveTime: 2**256-1,
                id:0
        });
        }
        
        return deviations[_assetName];
    }

    function getMedianData(bytes32 _assetName) external view whitelistCheck medianCheck returns (MedianData memory) {
        if (_assetName==USD){
            uint256[] memory medianValues = new uint256[](1);
            medianValues[0]=USD_PRICE;
            return MedianData ({
                name:USD, 
                values: medianValues,
                resolveTime: 2**256-1,
                id:0
            });
        }

        return medians[_assetName];
    }

    function getPrice(bytes32 _base, bytes32 _qoute) public view whitelistCheck returns (Price memory){
        Data memory base = _getPriceData(_base);
        Data memory quote = _getPriceData(_qoute);

        uint256 price = (base.value * 10**18) / quote.value;

        return Price({
            price: price,
            baseResolveTime: base.resolveTime,
            quoteResolveTime: quote.resolveTime
        });
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