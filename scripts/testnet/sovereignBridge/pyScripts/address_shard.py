import sys

from multiversx_sdk_core import Address, AddressComputer


def main():
    # input arguments
    address = Address.from_bech32(sys.argv[1])

    address_computer = AddressComputer()
    print(address_computer.get_shard_of_address(address))


if __name__ == "__main__":
    main()
