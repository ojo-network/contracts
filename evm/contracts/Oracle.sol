// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

contract PriceFeed is Ownable, AccessControl {
    error UnAuthorized();
    error MedianDisabled();
    address private _tempOwner;

    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");
    bytes32 public constant USD = bytes32("USD");
    uint256 constant USD_PRICE= 10**9;

    struct Data {
        bytes32 assetName;
        uint256 value;
        uint256 resolveTime;
        uint256 id;
    }

    struct MedianData{
        bytes32 assetName;
        uint256 resolveTime;
        uint256 id;
        uint256[] values;
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
    event WhitelistStatus(bool indexed status);
    event MedianStatus(bool indexed status);
    event RemovedFromWhitelist(address indexed user);

    constructor(address _relayer) {
        whitelist[_msgSender()]=true;
        whitelist[_relayer]=true;
        _grantRole(DEFAULT_ADMIN_ROLE, _msgSender());
        _grantRole(RELAYER_ROLE, _relayer);
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
        emit WhitelistStatus(_status);
    }

    function setMedianStatus(bool _status) external onlyOwner{
        medianDisabled=_status;
        emit MedianStatus(_status);
    }

    function postPrices(Data[] calldata _prices, bool _disableResolve) external onlyRole(RELAYER_ROLE) {
        uint256 time= _disableResolve?0:block.timestamp;
        for(uint256 i=0;i<_prices.length;i++){
            if (_prices[i].resolveTime>time){
                prices[_prices[i].assetName]= _prices[i];
            }
        }

        emit PricePosted(_msgSender(), block.timestamp);
    }

    function postDeviations(Data[] calldata _deviations,bool _disableResolve) external onlyRole(RELAYER_ROLE) {
        uint256 time = _disableResolve?0:block.timestamp;
        for(uint256 i=0;i<_deviations.length;i++){
            if (_deviations[i].resolveTime>time){
                deviations[_deviations[i].assetName]= _deviations[i];
            }
        }

        emit DeviationPosted(_msgSender(), block.timestamp);
    }

    function postMedians(MedianData[] calldata _medians,bool _disableResolve) external onlyRole(RELAYER_ROLE) {
        uint256 time = _disableResolve?0:block.timestamp;
        for(uint256 i=0;i<_medians.length;i++){
            if(_medians[i].resolveTime>time){
                medians[_medians[i].assetName]= _medians[i];
            }
        }

        emit MedianPosted(_msgSender(), block.timestamp);
    }

    function getPriceData(bytes32 _assetName) external view whitelistCheck returns (Data memory) {
        return _getPriceData(_assetName);
    }

    // solhint-disable-next-line max-line-length
    function getPriceDataBulk(bytes32[] calldata _assetNames) external view whitelistCheck returns (Data[] memory priceData) {
        priceData = new Data[](_assetNames.length);
        for (uint256 i = 0; i < _assetNames.length; i++) {
            priceData[i] = _getPriceData(_assetNames[i]);
        }

        return priceData;
    }

    function _getPriceData(bytes32 _assetName) internal view returns (Data memory){
        if (_assetName==USD){
            return  Data({
                assetName:USD,
                value: USD_PRICE,
                resolveTime: type(uint256).max,
                id:0
        });
        }

        return prices[_assetName];
    }

    function _getDeviationData(bytes32 _assetName) internal view returns (Data memory){
        if (_assetName==USD){
            return  Data({
                assetName:USD,
                value: 0,
                resolveTime: type(uint256).max,
                id:0
            });
        }

        return deviations[_assetName];
    }

    function getDeviationData(bytes32 _assetName) external view whitelistCheck returns (Data memory) {
        return _getDeviationData(_assetName);
    }

    // solhint-disable-next-line max-line-length
    function getDeviationDataBulk(bytes32[] calldata _assetNames) external view whitelistCheck returns (Data[] memory deviationData) {
        deviationData = new Data[](_assetNames.length);
        for (uint256 i = 0; i < _assetNames.length; i++) {
            deviationData[i] = _getDeviationData(_assetNames[i]);
        }

        return deviationData;
    }

    function _getMedianData(bytes32 _assetName) internal view returns (MedianData memory){
        if (_assetName == USD) {
            uint256[] memory medianValues = new uint256[](1);
            medianValues[0] = USD_PRICE;

            return MedianData({
                assetName: USD,
                values: medianValues,
                resolveTime: type(uint256).max,
                id: 0
            });
        }

        return medians[_assetName];
    }

    function getMedianData(bytes32 _assetName) external view whitelistCheck medianCheck returns (MedianData memory) {
        return _getMedianData(_assetName);
    }

    // solhint-disable-next-line max-line-length
    function getMedianDataBulk(bytes32[] calldata _assetNames) external view whitelistCheck medianCheck returns (MedianData[] memory medianData) {
        medianData = new MedianData[](_assetNames.length);
        for (uint256 i = 0; i < _assetNames.length; i++) {
            medianData[i] = _getMedianData(_assetNames[i]);
        }

        return medianData;
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

    // owner functions
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


    function transferOwnership(address _newOwner) override public onlyOwner{
        _tempOwner=_newOwner;
    }

    function claimOwnership() external{
        if(_msgSender()!=_tempOwner){
            revert UnAuthorized();
        }

        // remove previous owner role
        _revokeRole(DEFAULT_ADMIN_ROLE, owner());

        _grantRole(DEFAULT_ADMIN_ROLE, _tempOwner);

        _transferOwnership(_tempOwner);
    }
}
