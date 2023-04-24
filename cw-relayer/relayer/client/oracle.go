// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package client

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// PriceFeedData is an auto generated low-level Go binding around an user-defined struct.
type PriceFeedData struct {
	AssetName   [32]byte
	Value       *big.Int
	ResolveTime *big.Int
	Id          *big.Int
}

// PriceFeedMedianData is an auto generated low-level Go binding around an user-defined struct.
type PriceFeedMedianData struct {
	AssetName   [32]byte
	ResolveTime *big.Int
	Id          *big.Int
	Values      []*big.Int
}

// PriceFeedPrice is an auto generated low-level Go binding around an user-defined struct.
type PriceFeedPrice struct {
	Price            *big.Int
	BaseResolveTime  *big.Int
	QuoteResolveTime *big.Int
}

// OracleMetaData contains all meta data concerning the Oracle contract.
var OracleMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"MedianDisabled\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UnAuthorized\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"relayer\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"DeviationPosted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"relayer\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"MedianPosted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"MedianStatus\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"relayer\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"PricePosted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"RemovedFromWhitelist\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"WhitelistStatus\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"Whitelisted\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"RELAYER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"USD\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"relayer\",\"type\":\"address\"}],\"name\":\"assignRelayerRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_assetName\",\"type\":\"bytes32\"}],\"name\":\"getDeviationData\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"assetName\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"resolveTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"internalType\":\"structPriceFeed.Data\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"_assetNames\",\"type\":\"bytes32[]\"}],\"name\":\"getDeviationDataBulk\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"assetName\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"resolveTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"internalType\":\"structPriceFeed.Data[]\",\"name\":\"deviationData\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_assetName\",\"type\":\"bytes32\"}],\"name\":\"getMedianData\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"assetName\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"resolveTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"}],\"internalType\":\"structPriceFeed.MedianData\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"_assetNames\",\"type\":\"bytes32[]\"}],\"name\":\"getMedianDataBulk\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"assetName\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"resolveTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"}],\"internalType\":\"structPriceFeed.MedianData[]\",\"name\":\"medianData\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_base\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_qoute\",\"type\":\"bytes32\"}],\"name\":\"getPrice\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"baseResolveTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"quoteResolveTime\",\"type\":\"uint256\"}],\"internalType\":\"structPriceFeed.Price\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_assetName\",\"type\":\"bytes32\"}],\"name\":\"getPriceData\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"assetName\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"resolveTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"internalType\":\"structPriceFeed.Data\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"_assetNames\",\"type\":\"bytes32[]\"}],\"name\":\"getPriceDataBulk\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"assetName\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"resolveTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"internalType\":\"structPriceFeed.Data[]\",\"name\":\"priceData\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"assetName\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"resolveTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"internalType\":\"structPriceFeed.Data[]\",\"name\":\"_deviations\",\"type\":\"tuple[]\"},{\"internalType\":\"bool\",\"name\":\"_disableResolve\",\"type\":\"bool\"}],\"name\":\"postDeviations\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"assetName\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"resolveTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"}],\"internalType\":\"structPriceFeed.MedianData[]\",\"name\":\"_medians\",\"type\":\"tuple[]\"},{\"internalType\":\"bool\",\"name\":\"_disableResolve\",\"type\":\"bool\"}],\"name\":\"postMedians\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"assetName\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"resolveTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"internalType\":\"structPriceFeed.Data[]\",\"name\":\"_prices\",\"type\":\"tuple[]\"},{\"internalType\":\"bool\",\"name\":\"_disableResolve\",\"type\":\"bool\"}],\"name\":\"postPrices\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"removeAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"relayer\",\"type\":\"address\"}],\"name\":\"revokeRelayerRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"_status\",\"type\":\"bool\"}],\"name\":\"setMedianStatus\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"_status\",\"type\":\"bool\"}],\"name\":\"setWhitelistStatus\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"whitelistAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// OracleABI is the input ABI used to generate the binding from.
// Deprecated: Use OracleMetaData.ABI instead.
var OracleABI = OracleMetaData.ABI

// Oracle is an auto generated Go binding around an Ethereum contract.
type Oracle struct {
	OracleCaller     // Read-only binding to the contract
	OracleTransactor // Write-only binding to the contract
	OracleFilterer   // Log filterer for contract events
}

// OracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type OracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OracleSession struct {
	Contract     *Oracle           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OracleCallerSession struct {
	Contract *OracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// OracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OracleTransactorSession struct {
	Contract     *OracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type OracleRaw struct {
	Contract *Oracle // Generic contract binding to access the raw methods on
}

// OracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OracleCallerRaw struct {
	Contract *OracleCaller // Generic read-only contract binding to access the raw methods on
}

// OracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OracleTransactorRaw struct {
	Contract *OracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOracle creates a new instance of Oracle, bound to a specific deployed contract.
func NewOracle(address common.Address, backend bind.ContractBackend) (*Oracle, error) {
	contract, err := bindOracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Oracle{OracleCaller: OracleCaller{contract: contract}, OracleTransactor: OracleTransactor{contract: contract}, OracleFilterer: OracleFilterer{contract: contract}}, nil
}

// NewOracleCaller creates a new read-only instance of Oracle, bound to a specific deployed contract.
func NewOracleCaller(address common.Address, caller bind.ContractCaller) (*OracleCaller, error) {
	contract, err := bindOracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OracleCaller{contract: contract}, nil
}

// NewOracleTransactor creates a new write-only instance of Oracle, bound to a specific deployed contract.
func NewOracleTransactor(address common.Address, transactor bind.ContractTransactor) (*OracleTransactor, error) {
	contract, err := bindOracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OracleTransactor{contract: contract}, nil
}

// NewOracleFilterer creates a new log filterer instance of Oracle, bound to a specific deployed contract.
func NewOracleFilterer(address common.Address, filterer bind.ContractFilterer) (*OracleFilterer, error) {
	contract, err := bindOracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OracleFilterer{contract: contract}, nil
}

// bindOracle binds a generic wrapper to an already deployed contract.
func bindOracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OracleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Oracle *OracleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Oracle.Contract.OracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Oracle *OracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Oracle.Contract.OracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Oracle *OracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Oracle.Contract.OracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Oracle *OracleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Oracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Oracle *OracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Oracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Oracle *OracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Oracle.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Oracle *OracleCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Oracle *OracleSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Oracle.Contract.DEFAULTADMINROLE(&_Oracle.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Oracle *OracleCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Oracle.Contract.DEFAULTADMINROLE(&_Oracle.CallOpts)
}

// RELAYERROLE is a free data retrieval call binding the contract method 0x926d7d7f.
//
// Solidity: function RELAYER_ROLE() view returns(bytes32)
func (_Oracle *OracleCaller) RELAYERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "RELAYER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RELAYERROLE is a free data retrieval call binding the contract method 0x926d7d7f.
//
// Solidity: function RELAYER_ROLE() view returns(bytes32)
func (_Oracle *OracleSession) RELAYERROLE() ([32]byte, error) {
	return _Oracle.Contract.RELAYERROLE(&_Oracle.CallOpts)
}

// RELAYERROLE is a free data retrieval call binding the contract method 0x926d7d7f.
//
// Solidity: function RELAYER_ROLE() view returns(bytes32)
func (_Oracle *OracleCallerSession) RELAYERROLE() ([32]byte, error) {
	return _Oracle.Contract.RELAYERROLE(&_Oracle.CallOpts)
}

// USD is a free data retrieval call binding the contract method 0x1bf6c21b.
//
// Solidity: function USD() view returns(bytes32)
func (_Oracle *OracleCaller) USD(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "USD")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// USD is a free data retrieval call binding the contract method 0x1bf6c21b.
//
// Solidity: function USD() view returns(bytes32)
func (_Oracle *OracleSession) USD() ([32]byte, error) {
	return _Oracle.Contract.USD(&_Oracle.CallOpts)
}

// USD is a free data retrieval call binding the contract method 0x1bf6c21b.
//
// Solidity: function USD() view returns(bytes32)
func (_Oracle *OracleCallerSession) USD() ([32]byte, error) {
	return _Oracle.Contract.USD(&_Oracle.CallOpts)
}

// GetDeviationData is a free data retrieval call binding the contract method 0x437bb7da.
//
// Solidity: function getDeviationData(bytes32 _assetName) view returns((bytes32,uint256,uint256,uint256))
func (_Oracle *OracleCaller) GetDeviationData(opts *bind.CallOpts, _assetName [32]byte) (PriceFeedData, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getDeviationData", _assetName)

	if err != nil {
		return *new(PriceFeedData), err
	}

	out0 := *abi.ConvertType(out[0], new(PriceFeedData)).(*PriceFeedData)

	return out0, err

}

// GetDeviationData is a free data retrieval call binding the contract method 0x437bb7da.
//
// Solidity: function getDeviationData(bytes32 _assetName) view returns((bytes32,uint256,uint256,uint256))
func (_Oracle *OracleSession) GetDeviationData(_assetName [32]byte) (PriceFeedData, error) {
	return _Oracle.Contract.GetDeviationData(&_Oracle.CallOpts, _assetName)
}

// GetDeviationData is a free data retrieval call binding the contract method 0x437bb7da.
//
// Solidity: function getDeviationData(bytes32 _assetName) view returns((bytes32,uint256,uint256,uint256))
func (_Oracle *OracleCallerSession) GetDeviationData(_assetName [32]byte) (PriceFeedData, error) {
	return _Oracle.Contract.GetDeviationData(&_Oracle.CallOpts, _assetName)
}

// GetDeviationDataBulk is a free data retrieval call binding the contract method 0x51d0eb63.
//
// Solidity: function getDeviationDataBulk(bytes32[] _assetNames) view returns((bytes32,uint256,uint256,uint256)[] deviationData)
func (_Oracle *OracleCaller) GetDeviationDataBulk(opts *bind.CallOpts, _assetNames [][32]byte) ([]PriceFeedData, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getDeviationDataBulk", _assetNames)

	if err != nil {
		return *new([]PriceFeedData), err
	}

	out0 := *abi.ConvertType(out[0], new([]PriceFeedData)).(*[]PriceFeedData)

	return out0, err

}

// GetDeviationDataBulk is a free data retrieval call binding the contract method 0x51d0eb63.
//
// Solidity: function getDeviationDataBulk(bytes32[] _assetNames) view returns((bytes32,uint256,uint256,uint256)[] deviationData)
func (_Oracle *OracleSession) GetDeviationDataBulk(_assetNames [][32]byte) ([]PriceFeedData, error) {
	return _Oracle.Contract.GetDeviationDataBulk(&_Oracle.CallOpts, _assetNames)
}

// GetDeviationDataBulk is a free data retrieval call binding the contract method 0x51d0eb63.
//
// Solidity: function getDeviationDataBulk(bytes32[] _assetNames) view returns((bytes32,uint256,uint256,uint256)[] deviationData)
func (_Oracle *OracleCallerSession) GetDeviationDataBulk(_assetNames [][32]byte) ([]PriceFeedData, error) {
	return _Oracle.Contract.GetDeviationDataBulk(&_Oracle.CallOpts, _assetNames)
}

// GetMedianData is a free data retrieval call binding the contract method 0xb9fe8973.
//
// Solidity: function getMedianData(bytes32 _assetName) view returns((bytes32,uint256,uint256,uint256[]))
func (_Oracle *OracleCaller) GetMedianData(opts *bind.CallOpts, _assetName [32]byte) (PriceFeedMedianData, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getMedianData", _assetName)

	if err != nil {
		return *new(PriceFeedMedianData), err
	}

	out0 := *abi.ConvertType(out[0], new(PriceFeedMedianData)).(*PriceFeedMedianData)

	return out0, err

}

// GetMedianData is a free data retrieval call binding the contract method 0xb9fe8973.
//
// Solidity: function getMedianData(bytes32 _assetName) view returns((bytes32,uint256,uint256,uint256[]))
func (_Oracle *OracleSession) GetMedianData(_assetName [32]byte) (PriceFeedMedianData, error) {
	return _Oracle.Contract.GetMedianData(&_Oracle.CallOpts, _assetName)
}

// GetMedianData is a free data retrieval call binding the contract method 0xb9fe8973.
//
// Solidity: function getMedianData(bytes32 _assetName) view returns((bytes32,uint256,uint256,uint256[]))
func (_Oracle *OracleCallerSession) GetMedianData(_assetName [32]byte) (PriceFeedMedianData, error) {
	return _Oracle.Contract.GetMedianData(&_Oracle.CallOpts, _assetName)
}

// GetMedianDataBulk is a free data retrieval call binding the contract method 0x0fa1ae45.
//
// Solidity: function getMedianDataBulk(bytes32[] _assetNames) view returns((bytes32,uint256,uint256,uint256[])[] medianData)
func (_Oracle *OracleCaller) GetMedianDataBulk(opts *bind.CallOpts, _assetNames [][32]byte) ([]PriceFeedMedianData, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getMedianDataBulk", _assetNames)

	if err != nil {
		return *new([]PriceFeedMedianData), err
	}

	out0 := *abi.ConvertType(out[0], new([]PriceFeedMedianData)).(*[]PriceFeedMedianData)

	return out0, err

}

// GetMedianDataBulk is a free data retrieval call binding the contract method 0x0fa1ae45.
//
// Solidity: function getMedianDataBulk(bytes32[] _assetNames) view returns((bytes32,uint256,uint256,uint256[])[] medianData)
func (_Oracle *OracleSession) GetMedianDataBulk(_assetNames [][32]byte) ([]PriceFeedMedianData, error) {
	return _Oracle.Contract.GetMedianDataBulk(&_Oracle.CallOpts, _assetNames)
}

// GetMedianDataBulk is a free data retrieval call binding the contract method 0x0fa1ae45.
//
// Solidity: function getMedianDataBulk(bytes32[] _assetNames) view returns((bytes32,uint256,uint256,uint256[])[] medianData)
func (_Oracle *OracleCallerSession) GetMedianDataBulk(_assetNames [][32]byte) ([]PriceFeedMedianData, error) {
	return _Oracle.Contract.GetMedianDataBulk(&_Oracle.CallOpts, _assetNames)
}

// GetPrice is a free data retrieval call binding the contract method 0x0776f244.
//
// Solidity: function getPrice(bytes32 _base, bytes32 _qoute) view returns((uint256,uint256,uint256))
func (_Oracle *OracleCaller) GetPrice(opts *bind.CallOpts, _base [32]byte, _qoute [32]byte) (PriceFeedPrice, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getPrice", _base, _qoute)

	if err != nil {
		return *new(PriceFeedPrice), err
	}

	out0 := *abi.ConvertType(out[0], new(PriceFeedPrice)).(*PriceFeedPrice)

	return out0, err

}

// GetPrice is a free data retrieval call binding the contract method 0x0776f244.
//
// Solidity: function getPrice(bytes32 _base, bytes32 _qoute) view returns((uint256,uint256,uint256))
func (_Oracle *OracleSession) GetPrice(_base [32]byte, _qoute [32]byte) (PriceFeedPrice, error) {
	return _Oracle.Contract.GetPrice(&_Oracle.CallOpts, _base, _qoute)
}

// GetPrice is a free data retrieval call binding the contract method 0x0776f244.
//
// Solidity: function getPrice(bytes32 _base, bytes32 _qoute) view returns((uint256,uint256,uint256))
func (_Oracle *OracleCallerSession) GetPrice(_base [32]byte, _qoute [32]byte) (PriceFeedPrice, error) {
	return _Oracle.Contract.GetPrice(&_Oracle.CallOpts, _base, _qoute)
}

// GetPriceData is a free data retrieval call binding the contract method 0x43fa6211.
//
// Solidity: function getPriceData(bytes32 _assetName) view returns((bytes32,uint256,uint256,uint256))
func (_Oracle *OracleCaller) GetPriceData(opts *bind.CallOpts, _assetName [32]byte) (PriceFeedData, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getPriceData", _assetName)

	if err != nil {
		return *new(PriceFeedData), err
	}

	out0 := *abi.ConvertType(out[0], new(PriceFeedData)).(*PriceFeedData)

	return out0, err

}

// GetPriceData is a free data retrieval call binding the contract method 0x43fa6211.
//
// Solidity: function getPriceData(bytes32 _assetName) view returns((bytes32,uint256,uint256,uint256))
func (_Oracle *OracleSession) GetPriceData(_assetName [32]byte) (PriceFeedData, error) {
	return _Oracle.Contract.GetPriceData(&_Oracle.CallOpts, _assetName)
}

// GetPriceData is a free data retrieval call binding the contract method 0x43fa6211.
//
// Solidity: function getPriceData(bytes32 _assetName) view returns((bytes32,uint256,uint256,uint256))
func (_Oracle *OracleCallerSession) GetPriceData(_assetName [32]byte) (PriceFeedData, error) {
	return _Oracle.Contract.GetPriceData(&_Oracle.CallOpts, _assetName)
}

// GetPriceDataBulk is a free data retrieval call binding the contract method 0x9525eeda.
//
// Solidity: function getPriceDataBulk(bytes32[] _assetNames) view returns((bytes32,uint256,uint256,uint256)[] priceData)
func (_Oracle *OracleCaller) GetPriceDataBulk(opts *bind.CallOpts, _assetNames [][32]byte) ([]PriceFeedData, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getPriceDataBulk", _assetNames)

	if err != nil {
		return *new([]PriceFeedData), err
	}

	out0 := *abi.ConvertType(out[0], new([]PriceFeedData)).(*[]PriceFeedData)

	return out0, err

}

// GetPriceDataBulk is a free data retrieval call binding the contract method 0x9525eeda.
//
// Solidity: function getPriceDataBulk(bytes32[] _assetNames) view returns((bytes32,uint256,uint256,uint256)[] priceData)
func (_Oracle *OracleSession) GetPriceDataBulk(_assetNames [][32]byte) ([]PriceFeedData, error) {
	return _Oracle.Contract.GetPriceDataBulk(&_Oracle.CallOpts, _assetNames)
}

// GetPriceDataBulk is a free data retrieval call binding the contract method 0x9525eeda.
//
// Solidity: function getPriceDataBulk(bytes32[] _assetNames) view returns((bytes32,uint256,uint256,uint256)[] priceData)
func (_Oracle *OracleCallerSession) GetPriceDataBulk(_assetNames [][32]byte) ([]PriceFeedData, error) {
	return _Oracle.Contract.GetPriceDataBulk(&_Oracle.CallOpts, _assetNames)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Oracle *OracleCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Oracle *OracleSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Oracle.Contract.GetRoleAdmin(&_Oracle.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Oracle *OracleCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Oracle.Contract.GetRoleAdmin(&_Oracle.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Oracle *OracleCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Oracle *OracleSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Oracle.Contract.HasRole(&_Oracle.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Oracle *OracleCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Oracle.Contract.HasRole(&_Oracle.CallOpts, role, account)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Oracle *OracleCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Oracle *OracleSession) Owner() (common.Address, error) {
	return _Oracle.Contract.Owner(&_Oracle.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Oracle *OracleCallerSession) Owner() (common.Address, error) {
	return _Oracle.Contract.Owner(&_Oracle.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Oracle *OracleCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Oracle *OracleSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Oracle.Contract.SupportsInterface(&_Oracle.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Oracle *OracleCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Oracle.Contract.SupportsInterface(&_Oracle.CallOpts, interfaceId)
}

// AssignRelayerRole is a paid mutator transaction binding the contract method 0xb7dcbfa0.
//
// Solidity: function assignRelayerRole(address relayer) returns()
func (_Oracle *OracleTransactor) AssignRelayerRole(opts *bind.TransactOpts, relayer common.Address) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "assignRelayerRole", relayer)
}

// AssignRelayerRole is a paid mutator transaction binding the contract method 0xb7dcbfa0.
//
// Solidity: function assignRelayerRole(address relayer) returns()
func (_Oracle *OracleSession) AssignRelayerRole(relayer common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.AssignRelayerRole(&_Oracle.TransactOpts, relayer)
}

// AssignRelayerRole is a paid mutator transaction binding the contract method 0xb7dcbfa0.
//
// Solidity: function assignRelayerRole(address relayer) returns()
func (_Oracle *OracleTransactorSession) AssignRelayerRole(relayer common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.AssignRelayerRole(&_Oracle.TransactOpts, relayer)
}

// ClaimOwnership is a paid mutator transaction binding the contract method 0x4e71e0c8.
//
// Solidity: function claimOwnership() returns()
func (_Oracle *OracleTransactor) ClaimOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "claimOwnership")
}

// ClaimOwnership is a paid mutator transaction binding the contract method 0x4e71e0c8.
//
// Solidity: function claimOwnership() returns()
func (_Oracle *OracleSession) ClaimOwnership() (*types.Transaction, error) {
	return _Oracle.Contract.ClaimOwnership(&_Oracle.TransactOpts)
}

// ClaimOwnership is a paid mutator transaction binding the contract method 0x4e71e0c8.
//
// Solidity: function claimOwnership() returns()
func (_Oracle *OracleTransactorSession) ClaimOwnership() (*types.Transaction, error) {
	return _Oracle.Contract.ClaimOwnership(&_Oracle.TransactOpts)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Oracle *OracleTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Oracle *OracleSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.GrantRole(&_Oracle.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Oracle *OracleTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.GrantRole(&_Oracle.TransactOpts, role, account)
}

// PostDeviations is a paid mutator transaction binding the contract method 0x3bed995e.
//
// Solidity: function postDeviations((bytes32,uint256,uint256,uint256)[] _deviations, bool _disableResolve) returns()
func (_Oracle *OracleTransactor) PostDeviations(opts *bind.TransactOpts, _deviations []PriceFeedData, _disableResolve bool) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "postDeviations", _deviations, _disableResolve)
}

// PostDeviations is a paid mutator transaction binding the contract method 0x3bed995e.
//
// Solidity: function postDeviations((bytes32,uint256,uint256,uint256)[] _deviations, bool _disableResolve) returns()
func (_Oracle *OracleSession) PostDeviations(_deviations []PriceFeedData, _disableResolve bool) (*types.Transaction, error) {
	return _Oracle.Contract.PostDeviations(&_Oracle.TransactOpts, _deviations, _disableResolve)
}

// PostDeviations is a paid mutator transaction binding the contract method 0x3bed995e.
//
// Solidity: function postDeviations((bytes32,uint256,uint256,uint256)[] _deviations, bool _disableResolve) returns()
func (_Oracle *OracleTransactorSession) PostDeviations(_deviations []PriceFeedData, _disableResolve bool) (*types.Transaction, error) {
	return _Oracle.Contract.PostDeviations(&_Oracle.TransactOpts, _deviations, _disableResolve)
}

// PostMedians is a paid mutator transaction binding the contract method 0x3dd468f6.
//
// Solidity: function postMedians((bytes32,uint256,uint256,uint256[])[] _medians, bool _disableResolve) returns()
func (_Oracle *OracleTransactor) PostMedians(opts *bind.TransactOpts, _medians []PriceFeedMedianData, _disableResolve bool) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "postMedians", _medians, _disableResolve)
}

// PostMedians is a paid mutator transaction binding the contract method 0x3dd468f6.
//
// Solidity: function postMedians((bytes32,uint256,uint256,uint256[])[] _medians, bool _disableResolve) returns()
func (_Oracle *OracleSession) PostMedians(_medians []PriceFeedMedianData, _disableResolve bool) (*types.Transaction, error) {
	return _Oracle.Contract.PostMedians(&_Oracle.TransactOpts, _medians, _disableResolve)
}

// PostMedians is a paid mutator transaction binding the contract method 0x3dd468f6.
//
// Solidity: function postMedians((bytes32,uint256,uint256,uint256[])[] _medians, bool _disableResolve) returns()
func (_Oracle *OracleTransactorSession) PostMedians(_medians []PriceFeedMedianData, _disableResolve bool) (*types.Transaction, error) {
	return _Oracle.Contract.PostMedians(&_Oracle.TransactOpts, _medians, _disableResolve)
}

// PostPrices is a paid mutator transaction binding the contract method 0xa8bf524f.
//
// Solidity: function postPrices((bytes32,uint256,uint256,uint256)[] _prices, bool _disableResolve) returns()
func (_Oracle *OracleTransactor) PostPrices(opts *bind.TransactOpts, _prices []PriceFeedData, _disableResolve bool) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "postPrices", _prices, _disableResolve)
}

// PostPrices is a paid mutator transaction binding the contract method 0xa8bf524f.
//
// Solidity: function postPrices((bytes32,uint256,uint256,uint256)[] _prices, bool _disableResolve) returns()
func (_Oracle *OracleSession) PostPrices(_prices []PriceFeedData, _disableResolve bool) (*types.Transaction, error) {
	return _Oracle.Contract.PostPrices(&_Oracle.TransactOpts, _prices, _disableResolve)
}

// PostPrices is a paid mutator transaction binding the contract method 0xa8bf524f.
//
// Solidity: function postPrices((bytes32,uint256,uint256,uint256)[] _prices, bool _disableResolve) returns()
func (_Oracle *OracleTransactorSession) PostPrices(_prices []PriceFeedData, _disableResolve bool) (*types.Transaction, error) {
	return _Oracle.Contract.PostPrices(&_Oracle.TransactOpts, _prices, _disableResolve)
}

// RemoveAddress is a paid mutator transaction binding the contract method 0x4ba79dfe.
//
// Solidity: function removeAddress(address _user) returns()
func (_Oracle *OracleTransactor) RemoveAddress(opts *bind.TransactOpts, _user common.Address) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "removeAddress", _user)
}

// RemoveAddress is a paid mutator transaction binding the contract method 0x4ba79dfe.
//
// Solidity: function removeAddress(address _user) returns()
func (_Oracle *OracleSession) RemoveAddress(_user common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.RemoveAddress(&_Oracle.TransactOpts, _user)
}

// RemoveAddress is a paid mutator transaction binding the contract method 0x4ba79dfe.
//
// Solidity: function removeAddress(address _user) returns()
func (_Oracle *OracleTransactorSession) RemoveAddress(_user common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.RemoveAddress(&_Oracle.TransactOpts, _user)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Oracle *OracleTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Oracle *OracleSession) RenounceOwnership() (*types.Transaction, error) {
	return _Oracle.Contract.RenounceOwnership(&_Oracle.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Oracle *OracleTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Oracle.Contract.RenounceOwnership(&_Oracle.TransactOpts)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Oracle *OracleTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Oracle *OracleSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.RenounceRole(&_Oracle.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Oracle *OracleTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.RenounceRole(&_Oracle.TransactOpts, role, account)
}

// RevokeRelayerRole is a paid mutator transaction binding the contract method 0x142b7b37.
//
// Solidity: function revokeRelayerRole(address relayer) returns()
func (_Oracle *OracleTransactor) RevokeRelayerRole(opts *bind.TransactOpts, relayer common.Address) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "revokeRelayerRole", relayer)
}

// RevokeRelayerRole is a paid mutator transaction binding the contract method 0x142b7b37.
//
// Solidity: function revokeRelayerRole(address relayer) returns()
func (_Oracle *OracleSession) RevokeRelayerRole(relayer common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.RevokeRelayerRole(&_Oracle.TransactOpts, relayer)
}

// RevokeRelayerRole is a paid mutator transaction binding the contract method 0x142b7b37.
//
// Solidity: function revokeRelayerRole(address relayer) returns()
func (_Oracle *OracleTransactorSession) RevokeRelayerRole(relayer common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.RevokeRelayerRole(&_Oracle.TransactOpts, relayer)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Oracle *OracleTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Oracle *OracleSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.RevokeRole(&_Oracle.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Oracle *OracleTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.RevokeRole(&_Oracle.TransactOpts, role, account)
}

// SetMedianStatus is a paid mutator transaction binding the contract method 0xa0c639a3.
//
// Solidity: function setMedianStatus(bool _status) returns()
func (_Oracle *OracleTransactor) SetMedianStatus(opts *bind.TransactOpts, _status bool) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "setMedianStatus", _status)
}

// SetMedianStatus is a paid mutator transaction binding the contract method 0xa0c639a3.
//
// Solidity: function setMedianStatus(bool _status) returns()
func (_Oracle *OracleSession) SetMedianStatus(_status bool) (*types.Transaction, error) {
	return _Oracle.Contract.SetMedianStatus(&_Oracle.TransactOpts, _status)
}

// SetMedianStatus is a paid mutator transaction binding the contract method 0xa0c639a3.
//
// Solidity: function setMedianStatus(bool _status) returns()
func (_Oracle *OracleTransactorSession) SetMedianStatus(_status bool) (*types.Transaction, error) {
	return _Oracle.Contract.SetMedianStatus(&_Oracle.TransactOpts, _status)
}

// SetWhitelistStatus is a paid mutator transaction binding the contract method 0x4a999118.
//
// Solidity: function setWhitelistStatus(bool _status) returns()
func (_Oracle *OracleTransactor) SetWhitelistStatus(opts *bind.TransactOpts, _status bool) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "setWhitelistStatus", _status)
}

// SetWhitelistStatus is a paid mutator transaction binding the contract method 0x4a999118.
//
// Solidity: function setWhitelistStatus(bool _status) returns()
func (_Oracle *OracleSession) SetWhitelistStatus(_status bool) (*types.Transaction, error) {
	return _Oracle.Contract.SetWhitelistStatus(&_Oracle.TransactOpts, _status)
}

// SetWhitelistStatus is a paid mutator transaction binding the contract method 0x4a999118.
//
// Solidity: function setWhitelistStatus(bool _status) returns()
func (_Oracle *OracleTransactorSession) SetWhitelistStatus(_status bool) (*types.Transaction, error) {
	return _Oracle.Contract.SetWhitelistStatus(&_Oracle.TransactOpts, _status)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _newOwner) returns()
func (_Oracle *OracleTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _newOwner) returns()
func (_Oracle *OracleSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.TransferOwnership(&_Oracle.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _newOwner) returns()
func (_Oracle *OracleTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.TransferOwnership(&_Oracle.TransactOpts, _newOwner)
}

// WhitelistAddress is a paid mutator transaction binding the contract method 0x41566585.
//
// Solidity: function whitelistAddress(address _user) returns()
func (_Oracle *OracleTransactor) WhitelistAddress(opts *bind.TransactOpts, _user common.Address) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "whitelistAddress", _user)
}

// WhitelistAddress is a paid mutator transaction binding the contract method 0x41566585.
//
// Solidity: function whitelistAddress(address _user) returns()
func (_Oracle *OracleSession) WhitelistAddress(_user common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.WhitelistAddress(&_Oracle.TransactOpts, _user)
}

// WhitelistAddress is a paid mutator transaction binding the contract method 0x41566585.
//
// Solidity: function whitelistAddress(address _user) returns()
func (_Oracle *OracleTransactorSession) WhitelistAddress(_user common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.WhitelistAddress(&_Oracle.TransactOpts, _user)
}

// OracleDeviationPostedIterator is returned from FilterDeviationPosted and is used to iterate over the raw logs and unpacked data for DeviationPosted events raised by the Oracle contract.
type OracleDeviationPostedIterator struct {
	Event *OracleDeviationPosted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleDeviationPostedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleDeviationPosted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleDeviationPosted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleDeviationPostedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleDeviationPostedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleDeviationPosted represents a DeviationPosted event raised by the Oracle contract.
type OracleDeviationPosted struct {
	Relayer   common.Address
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDeviationPosted is a free log retrieval operation binding the contract event 0xd5d52731f0e3b46d9cb83a6dd7ec69a085960efc0b402cf42d2659937d3bacc2.
//
// Solidity: event DeviationPosted(address indexed relayer, uint256 indexed timestamp)
func (_Oracle *OracleFilterer) FilterDeviationPosted(opts *bind.FilterOpts, relayer []common.Address, timestamp []*big.Int) (*OracleDeviationPostedIterator, error) {

	var relayerRule []interface{}
	for _, relayerItem := range relayer {
		relayerRule = append(relayerRule, relayerItem)
	}
	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "DeviationPosted", relayerRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return &OracleDeviationPostedIterator{contract: _Oracle.contract, event: "DeviationPosted", logs: logs, sub: sub}, nil
}

// WatchDeviationPosted is a free log subscription operation binding the contract event 0xd5d52731f0e3b46d9cb83a6dd7ec69a085960efc0b402cf42d2659937d3bacc2.
//
// Solidity: event DeviationPosted(address indexed relayer, uint256 indexed timestamp)
func (_Oracle *OracleFilterer) WatchDeviationPosted(opts *bind.WatchOpts, sink chan<- *OracleDeviationPosted, relayer []common.Address, timestamp []*big.Int) (event.Subscription, error) {

	var relayerRule []interface{}
	for _, relayerItem := range relayer {
		relayerRule = append(relayerRule, relayerItem)
	}
	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "DeviationPosted", relayerRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleDeviationPosted)
				if err := _Oracle.contract.UnpackLog(event, "DeviationPosted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeviationPosted is a log parse operation binding the contract event 0xd5d52731f0e3b46d9cb83a6dd7ec69a085960efc0b402cf42d2659937d3bacc2.
//
// Solidity: event DeviationPosted(address indexed relayer, uint256 indexed timestamp)
func (_Oracle *OracleFilterer) ParseDeviationPosted(log types.Log) (*OracleDeviationPosted, error) {
	event := new(OracleDeviationPosted)
	if err := _Oracle.contract.UnpackLog(event, "DeviationPosted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleMedianPostedIterator is returned from FilterMedianPosted and is used to iterate over the raw logs and unpacked data for MedianPosted events raised by the Oracle contract.
type OracleMedianPostedIterator struct {
	Event *OracleMedianPosted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleMedianPostedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleMedianPosted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleMedianPosted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleMedianPostedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleMedianPostedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleMedianPosted represents a MedianPosted event raised by the Oracle contract.
type OracleMedianPosted struct {
	Relayer   common.Address
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMedianPosted is a free log retrieval operation binding the contract event 0xe0571ba3cefc33dc4d1277723989045ca18c1ae059a476162483af35284d98ea.
//
// Solidity: event MedianPosted(address indexed relayer, uint256 indexed timestamp)
func (_Oracle *OracleFilterer) FilterMedianPosted(opts *bind.FilterOpts, relayer []common.Address, timestamp []*big.Int) (*OracleMedianPostedIterator, error) {

	var relayerRule []interface{}
	for _, relayerItem := range relayer {
		relayerRule = append(relayerRule, relayerItem)
	}
	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "MedianPosted", relayerRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return &OracleMedianPostedIterator{contract: _Oracle.contract, event: "MedianPosted", logs: logs, sub: sub}, nil
}

// WatchMedianPosted is a free log subscription operation binding the contract event 0xe0571ba3cefc33dc4d1277723989045ca18c1ae059a476162483af35284d98ea.
//
// Solidity: event MedianPosted(address indexed relayer, uint256 indexed timestamp)
func (_Oracle *OracleFilterer) WatchMedianPosted(opts *bind.WatchOpts, sink chan<- *OracleMedianPosted, relayer []common.Address, timestamp []*big.Int) (event.Subscription, error) {

	var relayerRule []interface{}
	for _, relayerItem := range relayer {
		relayerRule = append(relayerRule, relayerItem)
	}
	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "MedianPosted", relayerRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleMedianPosted)
				if err := _Oracle.contract.UnpackLog(event, "MedianPosted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMedianPosted is a log parse operation binding the contract event 0xe0571ba3cefc33dc4d1277723989045ca18c1ae059a476162483af35284d98ea.
//
// Solidity: event MedianPosted(address indexed relayer, uint256 indexed timestamp)
func (_Oracle *OracleFilterer) ParseMedianPosted(log types.Log) (*OracleMedianPosted, error) {
	event := new(OracleMedianPosted)
	if err := _Oracle.contract.UnpackLog(event, "MedianPosted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleMedianStatusIterator is returned from FilterMedianStatus and is used to iterate over the raw logs and unpacked data for MedianStatus events raised by the Oracle contract.
type OracleMedianStatusIterator struct {
	Event *OracleMedianStatus // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleMedianStatusIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleMedianStatus)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleMedianStatus)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleMedianStatusIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleMedianStatusIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleMedianStatus represents a MedianStatus event raised by the Oracle contract.
type OracleMedianStatus struct {
	Status bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterMedianStatus is a free log retrieval operation binding the contract event 0x555063678b30a6267c8675e09f41f8ea73d45f04d314b0fe81abfd5523a9bcc9.
//
// Solidity: event MedianStatus(bool indexed status)
func (_Oracle *OracleFilterer) FilterMedianStatus(opts *bind.FilterOpts, status []bool) (*OracleMedianStatusIterator, error) {

	var statusRule []interface{}
	for _, statusItem := range status {
		statusRule = append(statusRule, statusItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "MedianStatus", statusRule)
	if err != nil {
		return nil, err
	}
	return &OracleMedianStatusIterator{contract: _Oracle.contract, event: "MedianStatus", logs: logs, sub: sub}, nil
}

// WatchMedianStatus is a free log subscription operation binding the contract event 0x555063678b30a6267c8675e09f41f8ea73d45f04d314b0fe81abfd5523a9bcc9.
//
// Solidity: event MedianStatus(bool indexed status)
func (_Oracle *OracleFilterer) WatchMedianStatus(opts *bind.WatchOpts, sink chan<- *OracleMedianStatus, status []bool) (event.Subscription, error) {

	var statusRule []interface{}
	for _, statusItem := range status {
		statusRule = append(statusRule, statusItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "MedianStatus", statusRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleMedianStatus)
				if err := _Oracle.contract.UnpackLog(event, "MedianStatus", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMedianStatus is a log parse operation binding the contract event 0x555063678b30a6267c8675e09f41f8ea73d45f04d314b0fe81abfd5523a9bcc9.
//
// Solidity: event MedianStatus(bool indexed status)
func (_Oracle *OracleFilterer) ParseMedianStatus(log types.Log) (*OracleMedianStatus, error) {
	event := new(OracleMedianStatus)
	if err := _Oracle.contract.UnpackLog(event, "MedianStatus", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Oracle contract.
type OracleOwnershipTransferredIterator struct {
	Event *OracleOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleOwnershipTransferred represents a OwnershipTransferred event raised by the Oracle contract.
type OracleOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Oracle *OracleFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*OracleOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &OracleOwnershipTransferredIterator{contract: _Oracle.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Oracle *OracleFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *OracleOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleOwnershipTransferred)
				if err := _Oracle.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Oracle *OracleFilterer) ParseOwnershipTransferred(log types.Log) (*OracleOwnershipTransferred, error) {
	event := new(OracleOwnershipTransferred)
	if err := _Oracle.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OraclePricePostedIterator is returned from FilterPricePosted and is used to iterate over the raw logs and unpacked data for PricePosted events raised by the Oracle contract.
type OraclePricePostedIterator struct {
	Event *OraclePricePosted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OraclePricePostedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OraclePricePosted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OraclePricePosted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OraclePricePostedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OraclePricePostedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OraclePricePosted represents a PricePosted event raised by the Oracle contract.
type OraclePricePosted struct {
	Relayer   common.Address
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPricePosted is a free log retrieval operation binding the contract event 0x1fd8aef6c0e3676e0208d3cd5bbb655c6b578f5157500b8c4fc1625ced29f676.
//
// Solidity: event PricePosted(address indexed relayer, uint256 indexed timestamp)
func (_Oracle *OracleFilterer) FilterPricePosted(opts *bind.FilterOpts, relayer []common.Address, timestamp []*big.Int) (*OraclePricePostedIterator, error) {

	var relayerRule []interface{}
	for _, relayerItem := range relayer {
		relayerRule = append(relayerRule, relayerItem)
	}
	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "PricePosted", relayerRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return &OraclePricePostedIterator{contract: _Oracle.contract, event: "PricePosted", logs: logs, sub: sub}, nil
}

// WatchPricePosted is a free log subscription operation binding the contract event 0x1fd8aef6c0e3676e0208d3cd5bbb655c6b578f5157500b8c4fc1625ced29f676.
//
// Solidity: event PricePosted(address indexed relayer, uint256 indexed timestamp)
func (_Oracle *OracleFilterer) WatchPricePosted(opts *bind.WatchOpts, sink chan<- *OraclePricePosted, relayer []common.Address, timestamp []*big.Int) (event.Subscription, error) {

	var relayerRule []interface{}
	for _, relayerItem := range relayer {
		relayerRule = append(relayerRule, relayerItem)
	}
	var timestampRule []interface{}
	for _, timestampItem := range timestamp {
		timestampRule = append(timestampRule, timestampItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "PricePosted", relayerRule, timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OraclePricePosted)
				if err := _Oracle.contract.UnpackLog(event, "PricePosted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePricePosted is a log parse operation binding the contract event 0x1fd8aef6c0e3676e0208d3cd5bbb655c6b578f5157500b8c4fc1625ced29f676.
//
// Solidity: event PricePosted(address indexed relayer, uint256 indexed timestamp)
func (_Oracle *OracleFilterer) ParsePricePosted(log types.Log) (*OraclePricePosted, error) {
	event := new(OraclePricePosted)
	if err := _Oracle.contract.UnpackLog(event, "PricePosted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleRemovedFromWhitelistIterator is returned from FilterRemovedFromWhitelist and is used to iterate over the raw logs and unpacked data for RemovedFromWhitelist events raised by the Oracle contract.
type OracleRemovedFromWhitelistIterator struct {
	Event *OracleRemovedFromWhitelist // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleRemovedFromWhitelistIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleRemovedFromWhitelist)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleRemovedFromWhitelist)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleRemovedFromWhitelistIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleRemovedFromWhitelistIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleRemovedFromWhitelist represents a RemovedFromWhitelist event raised by the Oracle contract.
type OracleRemovedFromWhitelist struct {
	User common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterRemovedFromWhitelist is a free log retrieval operation binding the contract event 0xcdd2e9b91a56913d370075169cefa1602ba36be5301664f752192bb1709df757.
//
// Solidity: event RemovedFromWhitelist(address indexed user)
func (_Oracle *OracleFilterer) FilterRemovedFromWhitelist(opts *bind.FilterOpts, user []common.Address) (*OracleRemovedFromWhitelistIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "RemovedFromWhitelist", userRule)
	if err != nil {
		return nil, err
	}
	return &OracleRemovedFromWhitelistIterator{contract: _Oracle.contract, event: "RemovedFromWhitelist", logs: logs, sub: sub}, nil
}

// WatchRemovedFromWhitelist is a free log subscription operation binding the contract event 0xcdd2e9b91a56913d370075169cefa1602ba36be5301664f752192bb1709df757.
//
// Solidity: event RemovedFromWhitelist(address indexed user)
func (_Oracle *OracleFilterer) WatchRemovedFromWhitelist(opts *bind.WatchOpts, sink chan<- *OracleRemovedFromWhitelist, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "RemovedFromWhitelist", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleRemovedFromWhitelist)
				if err := _Oracle.contract.UnpackLog(event, "RemovedFromWhitelist", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRemovedFromWhitelist is a log parse operation binding the contract event 0xcdd2e9b91a56913d370075169cefa1602ba36be5301664f752192bb1709df757.
//
// Solidity: event RemovedFromWhitelist(address indexed user)
func (_Oracle *OracleFilterer) ParseRemovedFromWhitelist(log types.Log) (*OracleRemovedFromWhitelist, error) {
	event := new(OracleRemovedFromWhitelist)
	if err := _Oracle.contract.UnpackLog(event, "RemovedFromWhitelist", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the Oracle contract.
type OracleRoleAdminChangedIterator struct {
	Event *OracleRoleAdminChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleRoleAdminChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleRoleAdminChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleRoleAdminChanged represents a RoleAdminChanged event raised by the Oracle contract.
type OracleRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Oracle *OracleFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*OracleRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &OracleRoleAdminChangedIterator{contract: _Oracle.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Oracle *OracleFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *OracleRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleRoleAdminChanged)
				if err := _Oracle.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Oracle *OracleFilterer) ParseRoleAdminChanged(log types.Log) (*OracleRoleAdminChanged, error) {
	event := new(OracleRoleAdminChanged)
	if err := _Oracle.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the Oracle contract.
type OracleRoleGrantedIterator struct {
	Event *OracleRoleGranted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleRoleGranted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleRoleGranted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleRoleGranted represents a RoleGranted event raised by the Oracle contract.
type OracleRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Oracle *OracleFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*OracleRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &OracleRoleGrantedIterator{contract: _Oracle.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Oracle *OracleFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *OracleRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleRoleGranted)
				if err := _Oracle.contract.UnpackLog(event, "RoleGranted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Oracle *OracleFilterer) ParseRoleGranted(log types.Log) (*OracleRoleGranted, error) {
	event := new(OracleRoleGranted)
	if err := _Oracle.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the Oracle contract.
type OracleRoleRevokedIterator struct {
	Event *OracleRoleRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleRoleRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleRoleRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleRoleRevoked represents a RoleRevoked event raised by the Oracle contract.
type OracleRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Oracle *OracleFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*OracleRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &OracleRoleRevokedIterator{contract: _Oracle.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Oracle *OracleFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *OracleRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleRoleRevoked)
				if err := _Oracle.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Oracle *OracleFilterer) ParseRoleRevoked(log types.Log) (*OracleRoleRevoked, error) {
	event := new(OracleRoleRevoked)
	if err := _Oracle.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleWhitelistStatusIterator is returned from FilterWhitelistStatus and is used to iterate over the raw logs and unpacked data for WhitelistStatus events raised by the Oracle contract.
type OracleWhitelistStatusIterator struct {
	Event *OracleWhitelistStatus // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleWhitelistStatusIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleWhitelistStatus)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleWhitelistStatus)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleWhitelistStatusIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleWhitelistStatusIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleWhitelistStatus represents a WhitelistStatus event raised by the Oracle contract.
type OracleWhitelistStatus struct {
	Status bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWhitelistStatus is a free log retrieval operation binding the contract event 0x7649162c113351f8435b6a5f0e731ebcfb2657f6eedc62a23254f382ac48f337.
//
// Solidity: event WhitelistStatus(bool indexed status)
func (_Oracle *OracleFilterer) FilterWhitelistStatus(opts *bind.FilterOpts, status []bool) (*OracleWhitelistStatusIterator, error) {

	var statusRule []interface{}
	for _, statusItem := range status {
		statusRule = append(statusRule, statusItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "WhitelistStatus", statusRule)
	if err != nil {
		return nil, err
	}
	return &OracleWhitelistStatusIterator{contract: _Oracle.contract, event: "WhitelistStatus", logs: logs, sub: sub}, nil
}

// WatchWhitelistStatus is a free log subscription operation binding the contract event 0x7649162c113351f8435b6a5f0e731ebcfb2657f6eedc62a23254f382ac48f337.
//
// Solidity: event WhitelistStatus(bool indexed status)
func (_Oracle *OracleFilterer) WatchWhitelistStatus(opts *bind.WatchOpts, sink chan<- *OracleWhitelistStatus, status []bool) (event.Subscription, error) {

	var statusRule []interface{}
	for _, statusItem := range status {
		statusRule = append(statusRule, statusItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "WhitelistStatus", statusRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleWhitelistStatus)
				if err := _Oracle.contract.UnpackLog(event, "WhitelistStatus", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWhitelistStatus is a log parse operation binding the contract event 0x7649162c113351f8435b6a5f0e731ebcfb2657f6eedc62a23254f382ac48f337.
//
// Solidity: event WhitelistStatus(bool indexed status)
func (_Oracle *OracleFilterer) ParseWhitelistStatus(log types.Log) (*OracleWhitelistStatus, error) {
	event := new(OracleWhitelistStatus)
	if err := _Oracle.contract.UnpackLog(event, "WhitelistStatus", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleWhitelistedIterator is returned from FilterWhitelisted and is used to iterate over the raw logs and unpacked data for Whitelisted events raised by the Oracle contract.
type OracleWhitelistedIterator struct {
	Event *OracleWhitelisted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleWhitelistedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleWhitelisted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleWhitelisted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleWhitelistedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleWhitelistedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleWhitelisted represents a Whitelisted event raised by the Oracle contract.
type OracleWhitelisted struct {
	User common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterWhitelisted is a free log retrieval operation binding the contract event 0xaab7954e9d246b167ef88aeddad35209ca2489d95a8aeb59e288d9b19fae5a54.
//
// Solidity: event Whitelisted(address indexed user)
func (_Oracle *OracleFilterer) FilterWhitelisted(opts *bind.FilterOpts, user []common.Address) (*OracleWhitelistedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "Whitelisted", userRule)
	if err != nil {
		return nil, err
	}
	return &OracleWhitelistedIterator{contract: _Oracle.contract, event: "Whitelisted", logs: logs, sub: sub}, nil
}

// WatchWhitelisted is a free log subscription operation binding the contract event 0xaab7954e9d246b167ef88aeddad35209ca2489d95a8aeb59e288d9b19fae5a54.
//
// Solidity: event Whitelisted(address indexed user)
func (_Oracle *OracleFilterer) WatchWhitelisted(opts *bind.WatchOpts, sink chan<- *OracleWhitelisted, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "Whitelisted", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleWhitelisted)
				if err := _Oracle.contract.UnpackLog(event, "Whitelisted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWhitelisted is a log parse operation binding the contract event 0xaab7954e9d246b167ef88aeddad35209ca2489d95a8aeb59e288d9b19fae5a54.
//
// Solidity: event Whitelisted(address indexed user)
func (_Oracle *OracleFilterer) ParseWhitelisted(log types.Log) (*OracleWhitelisted, error) {
	event := new(OracleWhitelisted)
	if err := _Oracle.contract.UnpackLog(event, "Whitelisted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
