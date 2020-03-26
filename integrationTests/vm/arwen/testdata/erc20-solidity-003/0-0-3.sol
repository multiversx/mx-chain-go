pragma solidity ^0.5.0;
contract Token {
	uint256 private totalSupply;
	string public name;
	string public symbol;
	mapping(address => uint256) public balances;

	event Transfer(address indexed _from, address indexed _to, uint256 _value);

	// Safemath
	function add(uint256 a, uint256 b) internal pure returns (uint256) {
		uint256 c = a + b;
		require(c >= a, "SafeMath: addition overflow");

		return c;
	}

	function sub(uint256 a, uint256 b) internal pure returns (uint256) {
		require(b <= a, "SafeMath: subtraction overflow");
		uint256 c = a - b;

		return c;
	}

	constructor() public {
		totalSupply = 100000000;
		name = "ERC20TokenDemo";
		symbol = "ETD";
		balances[msg.sender] = totalSupply;
	}

	function balanceOf(address account) view public returns (uint256) {
		return balances[account];
	}

	function transfer(address to, uint256 amount) public returns (bool) {
		balances[msg.sender] = sub(balances[msg.sender], amount);
		balances[to] = add(balances[to], amount);
		emit Transfer(msg.sender, to, amount);
		return true;
	}

	function () external payable {}
}
