import sys

from multiversx_sdk_core import Address, AddressComputer


def main():
    # input arguments
    address = Address.from_bech32(sys.argv[1])
    nonce = int(sys.argv[2])

    address_computer = AddressComputer()
    print(address_computer.compute_contract_address(address, nonce).to_bech32())


if __name__ == "__main__":
    main()
