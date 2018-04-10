package btc

import (
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil/txsort"
	"github.com/sirupsen/logrus"

	"github.com/shiftdevices/godbb/coins/btc/addresses"
	"github.com/shiftdevices/godbb/coins/btc/transactions"
	"github.com/shiftdevices/godbb/util/errp"
)

// SignTransaction signs all inputs. It assumes all outputs spent belong to this
// wallet. previousOutputs must contain all outputs which are spent by the transaction.
func SignTransaction(
	keyStore KeyStoreWithoutKeyDerivation,
	transaction *wire.MsgTx,
	previousOutputs map[wire.OutPoint]*transactions.TxOut,
	logEntry *logrus.Entry,
) error {
	logEntry.Info("Sign transaction")
	signatureHashes := [][]byte{}
	keyPaths := []string{}
	sigHashes := txscript.NewTxSigHashes(transaction)
	for index, txIn := range transaction.TxIn {
		spentOutput, ok := previousOutputs[txIn.PreviousOutPoint]
		if !ok {
			logEntry.Panic("output/input mismatch; there needs to be exactly one output being spent ber input")
			panic("output/input mismatch; there needs to be exactly one output being spent ber input")
		}
		address := spentOutput.Address.(*addresses.Address)
		isSegwit, subScript := address.SigHashData()
		var signatureHash []byte
		if isSegwit {
			var err error
			signatureHash, err = txscript.CalcWitnessSigHash(
				subScript, sigHashes, txscript.SigHashAll, transaction, index, spentOutput.Value)
			if err != nil {
				return errp.Wrap(err, "Failed to calculate SegWit signature hash")
			}
			logEntry.Debug("Calculated segwit signature hash")
		} else {
			var err error
			signatureHash, err = txscript.CalcSignatureHash(
				subScript, txscript.SigHashAll, transaction, index)
			if err != nil {
				return errp.Wrap(err, "Failed to calculate legacy signature hash")
			}
			logEntry.Debug("Calculated legacy signature hash")
		}

		signatureHashes = append(signatureHashes, signatureHash)
		keyPaths = append(keyPaths, spentOutput.Address.(*addresses.Address).KeyPath)
	}
	signatures, err := keyStore.Sign(signatureHashes, keyPaths)
	if err != nil {
		return errp.WithMessage(err, "Failed to sign signature hash")
	}
	if len(signatures) != len(transaction.TxIn) {
		panic("number of signatures doesn't match number of inputs")
	}
	for index, input := range transaction.TxIn {
		spentOutput := previousOutputs[input.PreviousOutPoint]
		address := spentOutput.Address.(*addresses.Address)
		signature := signatures[index]
		input.SignatureScript, input.Witness = address.InputData(signature)
	}
	// Sanity check: see if the created transaction is valid.
	if err := txValidityCheck(transaction, previousOutputs, sigHashes); err != nil {
		logEntry.WithField("error", err).Panic("Failed to pass transaction validity check")
		panic(err)
	}
	return nil
}

func txValidityCheck(transaction *wire.MsgTx, previousOutputs map[wire.OutPoint]*transactions.TxOut,
	sigHashes *txscript.TxSigHashes) error {
	if !txsort.IsSorted(transaction) {
		return errp.New("tx not bip69 conformant")
	}
	for index, txIn := range transaction.TxIn {
		spentOutput, ok := previousOutputs[txIn.PreviousOutPoint]
		if !ok {
			return errp.New("output/input mismatch; there needs to be exactly one output being spent per input")
		}
		engine, err := txscript.NewEngine(
			spentOutput.PkScript,
			transaction,
			index,
			txscript.StandardVerifyFlags, nil, sigHashes, spentOutput.Value)
		if err != nil {
			return errp.WithStack(err)
		}
		if err := engine.Execute(); err != nil {
			return errp.WithStack(err)
		}
	}
	return nil
}