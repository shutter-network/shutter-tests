// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package utils

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
	_ = abi.ConvertType
)

// DepositcontractMetaData contains all meta data concerning the Depositcontract contract.
var DepositcontractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"withdrawal_credentials\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"amount\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"index\",\"type\":\"bytes\"}],\"name\":\"DepositEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Paused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"Unpaused\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"pubkeys\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"withdrawal_credentials\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signatures\",\"type\":\"bytes\"},{\"internalType\":\"bytes32[]\",\"name\":\"deposit_data_roots\",\"type\":\"bytes32[]\"}],\"name\":\"batchDeposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"}],\"name\":\"claimTokens\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"claimWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_addresses\",\"type\":\"address[]\"}],\"name\":\"claimWithdrawals\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"withdrawal_credentials\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"deposit_data_root\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake_amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64[]\",\"name\":\"_amounts\",\"type\":\"uint64[]\"},{\"internalType\":\"address[]\",\"name\":\"_addresses\",\"type\":\"address[]\"}],\"name\":\"executeSystemWithdrawals\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint64[]\",\"name\":\"_amounts\",\"type\":\"uint64[]\"},{\"internalType\":\"address[]\",\"name\":\"_addresses\",\"type\":\"address[]\"}],\"name\":\"executeSystemWithdrawals\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"get_deposit_count\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"get_deposit_root\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"stake_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"onTokenTransfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stake_token\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unpause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIUnwrapper\",\"name\":\"_unwrapper\",\"type\":\"address\"},{\"internalType\":\"contractIERC20\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"unwrapTokens\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"validator_withdrawal_credentials\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"withdrawableAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// DepositcontractABI is the input ABI used to generate the binding from.
// Deprecated: Use DepositcontractMetaData.ABI instead.
var DepositcontractABI = DepositcontractMetaData.ABI

// Depositcontract is an auto generated Go binding around an Ethereum contract.
type Depositcontract struct {
	DepositcontractCaller     // Read-only binding to the contract
	DepositcontractTransactor // Write-only binding to the contract
	DepositcontractFilterer   // Log filterer for contract events
}

// DepositcontractCaller is an auto generated read-only Go binding around an Ethereum contract.
type DepositcontractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DepositcontractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DepositcontractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DepositcontractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DepositcontractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DepositcontractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DepositcontractSession struct {
	Contract     *Depositcontract  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DepositcontractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DepositcontractCallerSession struct {
	Contract *DepositcontractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// DepositcontractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DepositcontractTransactorSession struct {
	Contract     *DepositcontractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// DepositcontractRaw is an auto generated low-level Go binding around an Ethereum contract.
type DepositcontractRaw struct {
	Contract *Depositcontract // Generic contract binding to access the raw methods on
}

// DepositcontractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DepositcontractCallerRaw struct {
	Contract *DepositcontractCaller // Generic read-only contract binding to access the raw methods on
}

// DepositcontractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DepositcontractTransactorRaw struct {
	Contract *DepositcontractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDepositcontract creates a new instance of Depositcontract, bound to a specific deployed contract.
func NewDepositcontract(address common.Address, backend bind.ContractBackend) (*Depositcontract, error) {
	contract, err := bindDepositcontract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Depositcontract{DepositcontractCaller: DepositcontractCaller{contract: contract}, DepositcontractTransactor: DepositcontractTransactor{contract: contract}, DepositcontractFilterer: DepositcontractFilterer{contract: contract}}, nil
}

// NewDepositcontractCaller creates a new read-only instance of Depositcontract, bound to a specific deployed contract.
func NewDepositcontractCaller(address common.Address, caller bind.ContractCaller) (*DepositcontractCaller, error) {
	contract, err := bindDepositcontract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DepositcontractCaller{contract: contract}, nil
}

// NewDepositcontractTransactor creates a new write-only instance of Depositcontract, bound to a specific deployed contract.
func NewDepositcontractTransactor(address common.Address, transactor bind.ContractTransactor) (*DepositcontractTransactor, error) {
	contract, err := bindDepositcontract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DepositcontractTransactor{contract: contract}, nil
}

// NewDepositcontractFilterer creates a new log filterer instance of Depositcontract, bound to a specific deployed contract.
func NewDepositcontractFilterer(address common.Address, filterer bind.ContractFilterer) (*DepositcontractFilterer, error) {
	contract, err := bindDepositcontract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DepositcontractFilterer{contract: contract}, nil
}

// bindDepositcontract binds a generic wrapper to an already deployed contract.
func bindDepositcontract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := DepositcontractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Depositcontract *DepositcontractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Depositcontract.Contract.DepositcontractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Depositcontract *DepositcontractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Depositcontract.Contract.DepositcontractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Depositcontract *DepositcontractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Depositcontract.Contract.DepositcontractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Depositcontract *DepositcontractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Depositcontract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Depositcontract *DepositcontractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Depositcontract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Depositcontract *DepositcontractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Depositcontract.Contract.contract.Transact(opts, method, params...)
}

// GetDepositCount is a free data retrieval call binding the contract method 0x621fd130.
//
// Solidity: function get_deposit_count() view returns(bytes)
func (_Depositcontract *DepositcontractCaller) GetDepositCount(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _Depositcontract.contract.Call(opts, &out, "get_deposit_count")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetDepositCount is a free data retrieval call binding the contract method 0x621fd130.
//
// Solidity: function get_deposit_count() view returns(bytes)
func (_Depositcontract *DepositcontractSession) GetDepositCount() ([]byte, error) {
	return _Depositcontract.Contract.GetDepositCount(&_Depositcontract.CallOpts)
}

// GetDepositCount is a free data retrieval call binding the contract method 0x621fd130.
//
// Solidity: function get_deposit_count() view returns(bytes)
func (_Depositcontract *DepositcontractCallerSession) GetDepositCount() ([]byte, error) {
	return _Depositcontract.Contract.GetDepositCount(&_Depositcontract.CallOpts)
}

// GetDepositRoot is a free data retrieval call binding the contract method 0xc5f2892f.
//
// Solidity: function get_deposit_root() view returns(bytes32)
func (_Depositcontract *DepositcontractCaller) GetDepositRoot(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Depositcontract.contract.Call(opts, &out, "get_deposit_root")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetDepositRoot is a free data retrieval call binding the contract method 0xc5f2892f.
//
// Solidity: function get_deposit_root() view returns(bytes32)
func (_Depositcontract *DepositcontractSession) GetDepositRoot() ([32]byte, error) {
	return _Depositcontract.Contract.GetDepositRoot(&_Depositcontract.CallOpts)
}

// GetDepositRoot is a free data retrieval call binding the contract method 0xc5f2892f.
//
// Solidity: function get_deposit_root() view returns(bytes32)
func (_Depositcontract *DepositcontractCallerSession) GetDepositRoot() ([32]byte, error) {
	return _Depositcontract.Contract.GetDepositRoot(&_Depositcontract.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Depositcontract *DepositcontractCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Depositcontract.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Depositcontract *DepositcontractSession) Paused() (bool, error) {
	return _Depositcontract.Contract.Paused(&_Depositcontract.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_Depositcontract *DepositcontractCallerSession) Paused() (bool, error) {
	return _Depositcontract.Contract.Paused(&_Depositcontract.CallOpts)
}

// StakeToken is a free data retrieval call binding the contract method 0x640415bf.
//
// Solidity: function stake_token() view returns(address)
func (_Depositcontract *DepositcontractCaller) StakeToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Depositcontract.contract.Call(opts, &out, "stake_token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakeToken is a free data retrieval call binding the contract method 0x640415bf.
//
// Solidity: function stake_token() view returns(address)
func (_Depositcontract *DepositcontractSession) StakeToken() (common.Address, error) {
	return _Depositcontract.Contract.StakeToken(&_Depositcontract.CallOpts)
}

// StakeToken is a free data retrieval call binding the contract method 0x640415bf.
//
// Solidity: function stake_token() view returns(address)
func (_Depositcontract *DepositcontractCallerSession) StakeToken() (common.Address, error) {
	return _Depositcontract.Contract.StakeToken(&_Depositcontract.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_Depositcontract *DepositcontractCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Depositcontract.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_Depositcontract *DepositcontractSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Depositcontract.Contract.SupportsInterface(&_Depositcontract.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) pure returns(bool)
func (_Depositcontract *DepositcontractCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Depositcontract.Contract.SupportsInterface(&_Depositcontract.CallOpts, interfaceId)
}

// ValidatorWithdrawalCredentials is a free data retrieval call binding the contract method 0x24db4c46.
//
// Solidity: function validator_withdrawal_credentials(bytes ) view returns(bytes32)
func (_Depositcontract *DepositcontractCaller) ValidatorWithdrawalCredentials(opts *bind.CallOpts, arg0 []byte) ([32]byte, error) {
	var out []interface{}
	err := _Depositcontract.contract.Call(opts, &out, "validator_withdrawal_credentials", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ValidatorWithdrawalCredentials is a free data retrieval call binding the contract method 0x24db4c46.
//
// Solidity: function validator_withdrawal_credentials(bytes ) view returns(bytes32)
func (_Depositcontract *DepositcontractSession) ValidatorWithdrawalCredentials(arg0 []byte) ([32]byte, error) {
	return _Depositcontract.Contract.ValidatorWithdrawalCredentials(&_Depositcontract.CallOpts, arg0)
}

// ValidatorWithdrawalCredentials is a free data retrieval call binding the contract method 0x24db4c46.
//
// Solidity: function validator_withdrawal_credentials(bytes ) view returns(bytes32)
func (_Depositcontract *DepositcontractCallerSession) ValidatorWithdrawalCredentials(arg0 []byte) ([32]byte, error) {
	return _Depositcontract.Contract.ValidatorWithdrawalCredentials(&_Depositcontract.CallOpts, arg0)
}

// WithdrawableAmount is a free data retrieval call binding the contract method 0xbe7ab51b.
//
// Solidity: function withdrawableAmount(address ) view returns(uint256)
func (_Depositcontract *DepositcontractCaller) WithdrawableAmount(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Depositcontract.contract.Call(opts, &out, "withdrawableAmount", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WithdrawableAmount is a free data retrieval call binding the contract method 0xbe7ab51b.
//
// Solidity: function withdrawableAmount(address ) view returns(uint256)
func (_Depositcontract *DepositcontractSession) WithdrawableAmount(arg0 common.Address) (*big.Int, error) {
	return _Depositcontract.Contract.WithdrawableAmount(&_Depositcontract.CallOpts, arg0)
}

// WithdrawableAmount is a free data retrieval call binding the contract method 0xbe7ab51b.
//
// Solidity: function withdrawableAmount(address ) view returns(uint256)
func (_Depositcontract *DepositcontractCallerSession) WithdrawableAmount(arg0 common.Address) (*big.Int, error) {
	return _Depositcontract.Contract.WithdrawableAmount(&_Depositcontract.CallOpts, arg0)
}

// BatchDeposit is a paid mutator transaction binding the contract method 0xc82655b7.
//
// Solidity: function batchDeposit(bytes pubkeys, bytes withdrawal_credentials, bytes signatures, bytes32[] deposit_data_roots) returns()
func (_Depositcontract *DepositcontractTransactor) BatchDeposit(opts *bind.TransactOpts, pubkeys []byte, withdrawal_credentials []byte, signatures []byte, deposit_data_roots [][32]byte) (*types.Transaction, error) {
	return _Depositcontract.contract.Transact(opts, "batchDeposit", pubkeys, withdrawal_credentials, signatures, deposit_data_roots)
}

// BatchDeposit is a paid mutator transaction binding the contract method 0xc82655b7.
//
// Solidity: function batchDeposit(bytes pubkeys, bytes withdrawal_credentials, bytes signatures, bytes32[] deposit_data_roots) returns()
func (_Depositcontract *DepositcontractSession) BatchDeposit(pubkeys []byte, withdrawal_credentials []byte, signatures []byte, deposit_data_roots [][32]byte) (*types.Transaction, error) {
	return _Depositcontract.Contract.BatchDeposit(&_Depositcontract.TransactOpts, pubkeys, withdrawal_credentials, signatures, deposit_data_roots)
}

// BatchDeposit is a paid mutator transaction binding the contract method 0xc82655b7.
//
// Solidity: function batchDeposit(bytes pubkeys, bytes withdrawal_credentials, bytes signatures, bytes32[] deposit_data_roots) returns()
func (_Depositcontract *DepositcontractTransactorSession) BatchDeposit(pubkeys []byte, withdrawal_credentials []byte, signatures []byte, deposit_data_roots [][32]byte) (*types.Transaction, error) {
	return _Depositcontract.Contract.BatchDeposit(&_Depositcontract.TransactOpts, pubkeys, withdrawal_credentials, signatures, deposit_data_roots)
}

// ClaimTokens is a paid mutator transaction binding the contract method 0x69ffa08a.
//
// Solidity: function claimTokens(address _token, address _to) returns()
func (_Depositcontract *DepositcontractTransactor) ClaimTokens(opts *bind.TransactOpts, _token common.Address, _to common.Address) (*types.Transaction, error) {
	return _Depositcontract.contract.Transact(opts, "claimTokens", _token, _to)
}

// ClaimTokens is a paid mutator transaction binding the contract method 0x69ffa08a.
//
// Solidity: function claimTokens(address _token, address _to) returns()
func (_Depositcontract *DepositcontractSession) ClaimTokens(_token common.Address, _to common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.ClaimTokens(&_Depositcontract.TransactOpts, _token, _to)
}

// ClaimTokens is a paid mutator transaction binding the contract method 0x69ffa08a.
//
// Solidity: function claimTokens(address _token, address _to) returns()
func (_Depositcontract *DepositcontractTransactorSession) ClaimTokens(_token common.Address, _to common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.ClaimTokens(&_Depositcontract.TransactOpts, _token, _to)
}

// ClaimWithdrawal is a paid mutator transaction binding the contract method 0xa3066aab.
//
// Solidity: function claimWithdrawal(address _address) returns()
func (_Depositcontract *DepositcontractTransactor) ClaimWithdrawal(opts *bind.TransactOpts, _address common.Address) (*types.Transaction, error) {
	return _Depositcontract.contract.Transact(opts, "claimWithdrawal", _address)
}

// ClaimWithdrawal is a paid mutator transaction binding the contract method 0xa3066aab.
//
// Solidity: function claimWithdrawal(address _address) returns()
func (_Depositcontract *DepositcontractSession) ClaimWithdrawal(_address common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.ClaimWithdrawal(&_Depositcontract.TransactOpts, _address)
}

// ClaimWithdrawal is a paid mutator transaction binding the contract method 0xa3066aab.
//
// Solidity: function claimWithdrawal(address _address) returns()
func (_Depositcontract *DepositcontractTransactorSession) ClaimWithdrawal(_address common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.ClaimWithdrawal(&_Depositcontract.TransactOpts, _address)
}

// ClaimWithdrawals is a paid mutator transaction binding the contract method 0xbb30b8fd.
//
// Solidity: function claimWithdrawals(address[] _addresses) returns()
func (_Depositcontract *DepositcontractTransactor) ClaimWithdrawals(opts *bind.TransactOpts, _addresses []common.Address) (*types.Transaction, error) {
	return _Depositcontract.contract.Transact(opts, "claimWithdrawals", _addresses)
}

// ClaimWithdrawals is a paid mutator transaction binding the contract method 0xbb30b8fd.
//
// Solidity: function claimWithdrawals(address[] _addresses) returns()
func (_Depositcontract *DepositcontractSession) ClaimWithdrawals(_addresses []common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.ClaimWithdrawals(&_Depositcontract.TransactOpts, _addresses)
}

// ClaimWithdrawals is a paid mutator transaction binding the contract method 0xbb30b8fd.
//
// Solidity: function claimWithdrawals(address[] _addresses) returns()
func (_Depositcontract *DepositcontractTransactorSession) ClaimWithdrawals(_addresses []common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.ClaimWithdrawals(&_Depositcontract.TransactOpts, _addresses)
}

// Deposit is a paid mutator transaction binding the contract method 0x0cac9f31.
//
// Solidity: function deposit(bytes pubkey, bytes withdrawal_credentials, bytes signature, bytes32 deposit_data_root, uint256 stake_amount) returns()
func (_Depositcontract *DepositcontractTransactor) Deposit(opts *bind.TransactOpts, pubkey []byte, withdrawal_credentials []byte, signature []byte, deposit_data_root [32]byte, stake_amount *big.Int) (*types.Transaction, error) {
	return _Depositcontract.contract.Transact(opts, "deposit", pubkey, withdrawal_credentials, signature, deposit_data_root, stake_amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x0cac9f31.
//
// Solidity: function deposit(bytes pubkey, bytes withdrawal_credentials, bytes signature, bytes32 deposit_data_root, uint256 stake_amount) returns()
func (_Depositcontract *DepositcontractSession) Deposit(pubkey []byte, withdrawal_credentials []byte, signature []byte, deposit_data_root [32]byte, stake_amount *big.Int) (*types.Transaction, error) {
	return _Depositcontract.Contract.Deposit(&_Depositcontract.TransactOpts, pubkey, withdrawal_credentials, signature, deposit_data_root, stake_amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x0cac9f31.
//
// Solidity: function deposit(bytes pubkey, bytes withdrawal_credentials, bytes signature, bytes32 deposit_data_root, uint256 stake_amount) returns()
func (_Depositcontract *DepositcontractTransactorSession) Deposit(pubkey []byte, withdrawal_credentials []byte, signature []byte, deposit_data_root [32]byte, stake_amount *big.Int) (*types.Transaction, error) {
	return _Depositcontract.Contract.Deposit(&_Depositcontract.TransactOpts, pubkey, withdrawal_credentials, signature, deposit_data_root, stake_amount)
}

// ExecuteSystemWithdrawals is a paid mutator transaction binding the contract method 0x319ebe9c.
//
// Solidity: function executeSystemWithdrawals(uint64[] _amounts, address[] _addresses) returns()
func (_Depositcontract *DepositcontractTransactor) ExecuteSystemWithdrawals(opts *bind.TransactOpts, _amounts []uint64, _addresses []common.Address) (*types.Transaction, error) {
	return _Depositcontract.contract.Transact(opts, "executeSystemWithdrawals", _amounts, _addresses)
}

// ExecuteSystemWithdrawals is a paid mutator transaction binding the contract method 0x319ebe9c.
//
// Solidity: function executeSystemWithdrawals(uint64[] _amounts, address[] _addresses) returns()
func (_Depositcontract *DepositcontractSession) ExecuteSystemWithdrawals(_amounts []uint64, _addresses []common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.ExecuteSystemWithdrawals(&_Depositcontract.TransactOpts, _amounts, _addresses)
}

// ExecuteSystemWithdrawals is a paid mutator transaction binding the contract method 0x319ebe9c.
//
// Solidity: function executeSystemWithdrawals(uint64[] _amounts, address[] _addresses) returns()
func (_Depositcontract *DepositcontractTransactorSession) ExecuteSystemWithdrawals(_amounts []uint64, _addresses []common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.ExecuteSystemWithdrawals(&_Depositcontract.TransactOpts, _amounts, _addresses)
}

// ExecuteSystemWithdrawals0 is a paid mutator transaction binding the contract method 0x79d0c0bc.
//
// Solidity: function executeSystemWithdrawals(uint256 , uint64[] _amounts, address[] _addresses) returns()
func (_Depositcontract *DepositcontractTransactor) ExecuteSystemWithdrawals0(opts *bind.TransactOpts, arg0 *big.Int, _amounts []uint64, _addresses []common.Address) (*types.Transaction, error) {
	return _Depositcontract.contract.Transact(opts, "executeSystemWithdrawals0", arg0, _amounts, _addresses)
}

// ExecuteSystemWithdrawals0 is a paid mutator transaction binding the contract method 0x79d0c0bc.
//
// Solidity: function executeSystemWithdrawals(uint256 , uint64[] _amounts, address[] _addresses) returns()
func (_Depositcontract *DepositcontractSession) ExecuteSystemWithdrawals0(arg0 *big.Int, _amounts []uint64, _addresses []common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.ExecuteSystemWithdrawals0(&_Depositcontract.TransactOpts, arg0, _amounts, _addresses)
}

// ExecuteSystemWithdrawals0 is a paid mutator transaction binding the contract method 0x79d0c0bc.
//
// Solidity: function executeSystemWithdrawals(uint256 , uint64[] _amounts, address[] _addresses) returns()
func (_Depositcontract *DepositcontractTransactorSession) ExecuteSystemWithdrawals0(arg0 *big.Int, _amounts []uint64, _addresses []common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.ExecuteSystemWithdrawals0(&_Depositcontract.TransactOpts, arg0, _amounts, _addresses)
}

// OnTokenTransfer is a paid mutator transaction binding the contract method 0xa4c0ed36.
//
// Solidity: function onTokenTransfer(address , uint256 stake_amount, bytes data) returns(bool)
func (_Depositcontract *DepositcontractTransactor) OnTokenTransfer(opts *bind.TransactOpts, arg0 common.Address, stake_amount *big.Int, data []byte) (*types.Transaction, error) {
	return _Depositcontract.contract.Transact(opts, "onTokenTransfer", arg0, stake_amount, data)
}

// OnTokenTransfer is a paid mutator transaction binding the contract method 0xa4c0ed36.
//
// Solidity: function onTokenTransfer(address , uint256 stake_amount, bytes data) returns(bool)
func (_Depositcontract *DepositcontractSession) OnTokenTransfer(arg0 common.Address, stake_amount *big.Int, data []byte) (*types.Transaction, error) {
	return _Depositcontract.Contract.OnTokenTransfer(&_Depositcontract.TransactOpts, arg0, stake_amount, data)
}

// OnTokenTransfer is a paid mutator transaction binding the contract method 0xa4c0ed36.
//
// Solidity: function onTokenTransfer(address , uint256 stake_amount, bytes data) returns(bool)
func (_Depositcontract *DepositcontractTransactorSession) OnTokenTransfer(arg0 common.Address, stake_amount *big.Int, data []byte) (*types.Transaction, error) {
	return _Depositcontract.Contract.OnTokenTransfer(&_Depositcontract.TransactOpts, arg0, stake_amount, data)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Depositcontract *DepositcontractTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Depositcontract.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Depositcontract *DepositcontractSession) Pause() (*types.Transaction, error) {
	return _Depositcontract.Contract.Pause(&_Depositcontract.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_Depositcontract *DepositcontractTransactorSession) Pause() (*types.Transaction, error) {
	return _Depositcontract.Contract.Pause(&_Depositcontract.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Depositcontract *DepositcontractTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Depositcontract.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Depositcontract *DepositcontractSession) Unpause() (*types.Transaction, error) {
	return _Depositcontract.Contract.Unpause(&_Depositcontract.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_Depositcontract *DepositcontractTransactorSession) Unpause() (*types.Transaction, error) {
	return _Depositcontract.Contract.Unpause(&_Depositcontract.TransactOpts)
}

// UnwrapTokens is a paid mutator transaction binding the contract method 0x4694bd1e.
//
// Solidity: function unwrapTokens(address _unwrapper, address _token) returns()
func (_Depositcontract *DepositcontractTransactor) UnwrapTokens(opts *bind.TransactOpts, _unwrapper common.Address, _token common.Address) (*types.Transaction, error) {
	return _Depositcontract.contract.Transact(opts, "unwrapTokens", _unwrapper, _token)
}

// UnwrapTokens is a paid mutator transaction binding the contract method 0x4694bd1e.
//
// Solidity: function unwrapTokens(address _unwrapper, address _token) returns()
func (_Depositcontract *DepositcontractSession) UnwrapTokens(_unwrapper common.Address, _token common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.UnwrapTokens(&_Depositcontract.TransactOpts, _unwrapper, _token)
}

// UnwrapTokens is a paid mutator transaction binding the contract method 0x4694bd1e.
//
// Solidity: function unwrapTokens(address _unwrapper, address _token) returns()
func (_Depositcontract *DepositcontractTransactorSession) UnwrapTokens(_unwrapper common.Address, _token common.Address) (*types.Transaction, error) {
	return _Depositcontract.Contract.UnwrapTokens(&_Depositcontract.TransactOpts, _unwrapper, _token)
}

// DepositcontractDepositEventIterator is returned from FilterDepositEvent and is used to iterate over the raw logs and unpacked data for DepositEvent events raised by the Depositcontract contract.
type DepositcontractDepositEventIterator struct {
	Event *DepositcontractDepositEvent // Event containing the contract specifics and raw log

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
func (it *DepositcontractDepositEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DepositcontractDepositEvent)
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
		it.Event = new(DepositcontractDepositEvent)
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
func (it *DepositcontractDepositEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DepositcontractDepositEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DepositcontractDepositEvent represents a DepositEvent event raised by the Depositcontract contract.
type DepositcontractDepositEvent struct {
	Pubkey                []byte
	WithdrawalCredentials []byte
	Amount                []byte
	Signature             []byte
	Index                 []byte
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterDepositEvent is a free log retrieval operation binding the contract event 0x649bbc62d0e31342afea4e5cd82d4049e7e1ee912fc0889aa790803be39038c5.
//
// Solidity: event DepositEvent(bytes pubkey, bytes withdrawal_credentials, bytes amount, bytes signature, bytes index)
func (_Depositcontract *DepositcontractFilterer) FilterDepositEvent(opts *bind.FilterOpts) (*DepositcontractDepositEventIterator, error) {

	logs, sub, err := _Depositcontract.contract.FilterLogs(opts, "DepositEvent")
	if err != nil {
		return nil, err
	}
	return &DepositcontractDepositEventIterator{contract: _Depositcontract.contract, event: "DepositEvent", logs: logs, sub: sub}, nil
}

// WatchDepositEvent is a free log subscription operation binding the contract event 0x649bbc62d0e31342afea4e5cd82d4049e7e1ee912fc0889aa790803be39038c5.
//
// Solidity: event DepositEvent(bytes pubkey, bytes withdrawal_credentials, bytes amount, bytes signature, bytes index)
func (_Depositcontract *DepositcontractFilterer) WatchDepositEvent(opts *bind.WatchOpts, sink chan<- *DepositcontractDepositEvent) (event.Subscription, error) {

	logs, sub, err := _Depositcontract.contract.WatchLogs(opts, "DepositEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DepositcontractDepositEvent)
				if err := _Depositcontract.contract.UnpackLog(event, "DepositEvent", log); err != nil {
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

// ParseDepositEvent is a log parse operation binding the contract event 0x649bbc62d0e31342afea4e5cd82d4049e7e1ee912fc0889aa790803be39038c5.
//
// Solidity: event DepositEvent(bytes pubkey, bytes withdrawal_credentials, bytes amount, bytes signature, bytes index)
func (_Depositcontract *DepositcontractFilterer) ParseDepositEvent(log types.Log) (*DepositcontractDepositEvent, error) {
	event := new(DepositcontractDepositEvent)
	if err := _Depositcontract.contract.UnpackLog(event, "DepositEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DepositcontractPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the Depositcontract contract.
type DepositcontractPausedIterator struct {
	Event *DepositcontractPaused // Event containing the contract specifics and raw log

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
func (it *DepositcontractPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DepositcontractPaused)
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
		it.Event = new(DepositcontractPaused)
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
func (it *DepositcontractPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DepositcontractPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DepositcontractPaused represents a Paused event raised by the Depositcontract contract.
type DepositcontractPaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Depositcontract *DepositcontractFilterer) FilterPaused(opts *bind.FilterOpts) (*DepositcontractPausedIterator, error) {

	logs, sub, err := _Depositcontract.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &DepositcontractPausedIterator{contract: _Depositcontract.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Depositcontract *DepositcontractFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *DepositcontractPaused) (event.Subscription, error) {

	logs, sub, err := _Depositcontract.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DepositcontractPaused)
				if err := _Depositcontract.contract.UnpackLog(event, "Paused", log); err != nil {
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

// ParsePaused is a log parse operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_Depositcontract *DepositcontractFilterer) ParsePaused(log types.Log) (*DepositcontractPaused, error) {
	event := new(DepositcontractPaused)
	if err := _Depositcontract.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DepositcontractUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the Depositcontract contract.
type DepositcontractUnpausedIterator struct {
	Event *DepositcontractUnpaused // Event containing the contract specifics and raw log

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
func (it *DepositcontractUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DepositcontractUnpaused)
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
		it.Event = new(DepositcontractUnpaused)
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
func (it *DepositcontractUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DepositcontractUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DepositcontractUnpaused represents a Unpaused event raised by the Depositcontract contract.
type DepositcontractUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Depositcontract *DepositcontractFilterer) FilterUnpaused(opts *bind.FilterOpts) (*DepositcontractUnpausedIterator, error) {

	logs, sub, err := _Depositcontract.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &DepositcontractUnpausedIterator{contract: _Depositcontract.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Depositcontract *DepositcontractFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *DepositcontractUnpaused) (event.Subscription, error) {

	logs, sub, err := _Depositcontract.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DepositcontractUnpaused)
				if err := _Depositcontract.contract.UnpackLog(event, "Unpaused", log); err != nil {
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

// ParseUnpaused is a log parse operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_Depositcontract *DepositcontractFilterer) ParseUnpaused(log types.Log) (*DepositcontractUnpaused, error) {
	event := new(DepositcontractUnpaused)
	if err := _Depositcontract.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
