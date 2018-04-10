package addresses

import (
	"fmt"

	"github.com/shiftdevices/godbb/util/errp"
	"github.com/sirupsen/logrus"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
)

// AddressChain manages a chain of addresses derived from an xpub.
type AddressChain struct {
	xpub        *hdkeychain.ExtendedKey
	net         *chaincfg.Params
	gapLimit    int
	chainIndex  uint32
	addressType AddressType
	addresses   []*Address
	logEntry    *logrus.Entry
}

// NewAddressChain creates an address chain starting at m/<chainIndex> from the given xpub. xpub
// must be public (neutered) and the xpub type must match the passed net.
func NewAddressChain(
	xpub *hdkeychain.ExtendedKey,
	net *chaincfg.Params,
	gapLimit int,
	chainIndex uint32,
	addressType AddressType,
	logEntry *logrus.Entry,
) *AddressChain {
	if xpub.IsPrivate() {
		panic("Extended key is private! Only public keys are accepted")
	}
	if !xpub.IsForNet(net) {
		panic(errp.New("xpub does not match provided net"))
	}
	chainXPub, err := xpub.Child(chainIndex)
	if err != nil {
		logEntry.WithField("error", err).WithError(err)
		panic(err)
	}
	return &AddressChain{
		xpub:        chainXPub,
		net:         net,
		gapLimit:    gapLimit,
		chainIndex:  chainIndex,
		addressType: addressType,
		addresses:   []*Address{},
		logEntry: logEntry.WithFields(logrus.Fields{"group": "addresses", "net": net.Name,
			"gap-limit": gapLimit, "address-type": addressType}),
	}
}

// GetUnused returns the first unused address. EnsureAddresses() must be called beforehand.
func (addresses *AddressChain) GetUnused() *Address {
	if addresses.unusedTailCount() != addresses.gapLimit {
		addresses.logEntry.Panic("Concurrency error: Addresses not synced correctly")
		panic("concurrency error; addresses not synced correctly")
	}
	return addresses.addresses[len(addresses.addresses)-addresses.gapLimit]
}

func (addresses *AddressChain) getPubKey(index uint32) *btcec.PublicKey {
	addresses.logEntry.Debug("Get public key")
	xpub, err := addresses.xpub.Child(index)
	if err != nil {
		addresses.logEntry.WithFields(logrus.Fields{"index": index, "error": err}).
			Panic("Failed to get XPub child")
		panic(err)
	}
	publicKey, err := xpub.ECPubKey()
	if err != nil {
		addresses.logEntry.WithField("error", err).Panic("Failed to get EC pubkey")
		panic(err)
	}
	return publicKey
}

// addAddress appends a new address at the end of the chain.
func (addresses *AddressChain) addAddress() *Address {
	addresses.logEntry.Debug("Add new address to chain")
	index := len(addresses.addresses)
	publicKey := addresses.getPubKey(uint32(index))
	addressWithPK := NewAddress(
		publicKey,
		addresses.net,
		fmt.Sprintf("%d/%d", addresses.chainIndex, index),
		addresses.addressType,
		addresses.logEntry,
	)
	addresses.addresses = append(addresses.addresses, addressWithPK)
	return addressWithPK

}

// unusedTailCount returns the number of unused addresses at the end of the chain.
func (addresses *AddressChain) unusedTailCount() int {
	count := 0
	for i := len(addresses.addresses) - 1; i >= 0; i-- {
		if addresses.addresses[i].isUsed() {
			break
		}
		count++
	}
	addresses.logEntry.WithField("tail-count", count).Debug("Unused tail count")
	return count
}

// Contains returns whether the address is part of the address chain.
func (addresses *AddressChain) Contains(checkAddress btcutil.Address) bool {
	// todo: add map for constant time lookup
	for _, address := range addresses.addresses {
		if checkAddress.String() == address.String() {
			return true
		}
	}
	return false
}

// EnsureAddresses appends addresses to the address chain until there are `gapLimit` unused unused
// ones, and returns the new addresses.
func (addresses *AddressChain) EnsureAddresses() []*Address {
	addedAddresses := []*Address{}
	unusedAddressCount := addresses.unusedTailCount()
	for i := 0; i < addresses.gapLimit-unusedAddressCount; i++ {
		addedAddresses = append(addedAddresses, addresses.addAddress())
	}
	return addedAddresses
}