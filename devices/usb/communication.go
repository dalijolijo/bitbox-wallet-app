package usb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"sync"
	"unicode"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
	"github.com/shiftdevices/godbb/devices/bitbox"
	"github.com/shiftdevices/godbb/util/aes"
	"github.com/shiftdevices/godbb/util/errp"
	"github.com/shiftdevices/godbb/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	usbReportSize = 64
	hwwCID        = 0xff000000
	// initial frame identifier
	u2fHIDTypeInit = 0x80
	// first vendor defined command
	u2fHIDVendorFirst = u2fHIDTypeInit | 0x40
	hwwCMD            = u2fHIDVendorFirst | 0x01
)

// Communication encodes JSON messages to/from a bitbox. The serialized messages are sent/received
// as USB packets, following the ISO 7816-4 standard.
type Communication struct {
	device   io.ReadWriteCloser
	mutex    sync.Mutex
	logEntry *logrus.Entry
}

// NewCommunication creates a new Communication.
func NewCommunication(device io.ReadWriteCloser) *Communication {
	return &Communication{
		device:   device,
		mutex:    sync.Mutex{},
		logEntry: logging.Log.WithGroup("usb"),
	}
}

// Close closes the underlying device.
func (communication *Communication) Close() {
	if err := communication.device.Close(); err != nil {
		communication.logEntry.WithField("error", err).Panic(err)
		panic(err)
	}
}

func (communication *Communication) sendFrame(msg string) error {
	dataLen := len(msg)
	if dataLen == 0 {
		return nil
	}
	send := func(header []byte, readFrom *bytes.Buffer) error {
		buf := new(bytes.Buffer)
		buf.Write(header)
		buf.Write(readFrom.Next(usbReportSize - buf.Len()))
		for buf.Len() < usbReportSize {
			buf.WriteByte(0xee)
		}
		_, err := communication.device.Write(buf.Bytes())
		return errors.WithMessage(errors.WithStack(err), "Failed to send message")
	}
	readBuffer := bytes.NewBuffer([]byte(msg))
	// init frame
	header := new(bytes.Buffer)
	if err := binary.Write(header, binary.BigEndian, uint32(hwwCID)); err != nil {
		return errp.WithStack(err)
	}
	if err := binary.Write(header, binary.BigEndian, uint8(hwwCMD)); err != nil {
		return errp.WithStack(err)
	}
	if err := binary.Write(header, binary.BigEndian, uint16(dataLen&0xFFFF)); err != nil {
		return errp.WithStack(err)
	}
	if err := send(header.Bytes(), readBuffer); err != nil {
		return err
	}
	for seq := 0; readBuffer.Len() > 0; seq++ {
		// cont frame
		header = new(bytes.Buffer)
		if err := binary.Write(header, binary.BigEndian, uint32(hwwCID)); err != nil {
			return errp.WithStack(err)
		}
		if err := binary.Write(header, binary.BigEndian, uint8(seq)); err != nil {
			return errp.WithStack(err)
		}
		if err := send(header.Bytes(), readBuffer); err != nil {
			return err
		}
	}
	return nil
}

func (communication *Communication) readFrame() ([]byte, error) {
	read := make([]byte, usbReportSize)
	readLen, err := communication.device.Read(read)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if readLen < 7 {
		return nil, errors.New("expected minimum read length of 7")
	}
	if read[0] != 0xff || read[1] != 0 || read[2] != 0 || read[3] != 0 {
		return nil, errors.New("USB command ID mismatch")
	}
	if read[4] != hwwCMD {
		return nil, errp.Newf("USB command frame mismatch (%d, expected %d)", read[4], hwwCMD)
	}
	data := new(bytes.Buffer)
	dataLen := int(read[5])*256 + int(read[6])
	data.Write(read[7:readLen])
	idx := len(read) - 7
	for idx < dataLen {
		readLen, err = communication.device.Read(read)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if readLen < 5 {
			return nil, errors.New("expected minimum read length of 7")
		}
		data.Write(read[5:readLen])
		idx += readLen - 5
	}
	return data.Bytes(), nil
}

// SendBootloader sends a message in the format the bootloader expects and fetches the response.
func (communication *Communication) SendBootloader(msg []byte) ([]byte, error) {
	communication.mutex.Lock()
	defer communication.mutex.Unlock()
	const (
		maxSendLen = 4098
		maxReadLen = 256
	)
	if len(msg) > maxSendLen {
		communication.logEntry.WithFields(logrus.Fields{"message-length": len(msg),
			"max-send-length": maxSendLen}).Panic("Message too long")
		panic("message too long")
	}
	var buf bytes.Buffer
	buf.WriteByte(0)
	buf.Write(msg)
	buf.Write(bytes.Repeat([]byte{0}, maxSendLen-len(msg)))
	_, err := communication.device.Write(buf.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var read bytes.Buffer
	for read.Len() < maxReadLen {
		currentRead := make([]byte, maxReadLen)
		readLen, err := communication.device.Read(currentRead)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		read.Write(currentRead[:readLen])
	}
	return bytes.TrimRight(read.Bytes(), "\x00\t\r\n"), nil
}

func hideValues(cmd map[string]interface{}) {
	for k, v := range cmd {
		value, ok := v.(map[string]interface{})
		if ok {
			hideValues(value)
		} else {
			cmd[k] = "****"
		}
	}
}

func logCensoredCmd(logEntry *logrus.Entry, msg string, receiving bool) error {
	if logging.Log.Level >= logrus.DebugLevel {
		cmd := map[string]interface{}{}
		if err := json.Unmarshal([]byte(msg), &cmd); err != nil {
			return errp.New("Failed to unmarshal message")
		}
		hideValues(cmd)
		censoredMsg, err := json.Marshal(cmd)
		if err != nil {
			logEntry.WithField("error", err).Error("Failed to censor message")
		} else {
			direction := "Sending"
			if receiving {
				direction = "Receiving"
			}
			logEntry.WithField("msg", string(censoredMsg)).Infof("%s message", direction)
		}
	}
	return nil
}

// SendPlain sends an unecrypted message. The response is json-deserialized into a map.
func (communication *Communication) SendPlain(msg string) (map[string]interface{}, error) {
	if err := logCensoredCmd(communication.logEntry, msg, false); err != nil {
		communication.logEntry.WithField("msg", msg).Debug("Sending (encrypted) command")
	}
	communication.mutex.Lock()
	defer communication.mutex.Unlock()
	if err := communication.sendFrame(msg); err != nil {
		return nil, err
	}
	reply, err := communication.readFrame()
	if err != nil {
		return nil, err
	}
	reply = bytes.TrimRightFunc(reply, func(r rune) bool { return unicode.IsSpace(r) || r == 0 })
	err = logCensoredCmd(communication.logEntry, string(reply), true)
	if err != nil {
		return nil, errp.WithContext(err, errp.Context{"reply": string(reply)})
	}
	jsonResult := map[string]interface{}{}
	if err := json.Unmarshal(reply, &jsonResult); err != nil {
		return nil, errp.Wrap(err, "Failed to unmarshal reply")
	}
	if err := maybeDBBErr(jsonResult); err != nil {
		return nil, err
	}
	return jsonResult, nil
}

func maybeDBBErr(jsonResult map[string]interface{}) error {
	if errMap, ok := jsonResult["error"].(map[string]interface{}); ok {
		errMsg, ok := errMap["message"].(string)
		if !ok {
			return errp.WithContext(errp.New("Unexpected reply"), errp.Context{"reply": errMap})
		}
		errCode, ok := errMap["code"].(float64)
		if !ok {
			return errp.WithContext(errp.New("Unexpected reply"), errp.Context{"reply": errMap})
		}
		return bitbox.NewError(errMsg, errCode)
	}
	return nil
}

// SendEncrypt sends an encrypted message. The response is json-deserialized into a map. If the
// response contains an error field, it is returned as a DBBErr.
func (communication *Communication) SendEncrypt(msg, password string) (map[string]interface{}, error) {
	if err := logCensoredCmd(communication.logEntry, msg, false); err != nil {
		return nil, errp.WithMessage(err, "Invalid JSON passed. Continuing anyway")
	}
	secret := chainhash.DoubleHashB([]byte(password))
	cipherText, err := aes.Encrypt(secret, []byte(msg))
	if err != nil {
		return nil, errp.WithMessage(err, "Failed to encrypt command")
	}
	jsonResult, err := communication.SendPlain(cipherText)
	if err != nil {
		return nil, errp.WithMessage(err, "Failed to send cipher text")
	}
	if cipherText, ok := jsonResult["ciphertext"].(string); ok {
		plainText, err := aes.Decrypt(secret, cipherText)
		if err != nil {
			return nil, errp.WithMessage(err, "Failed to decrypt reply")
		}
		err = logCensoredCmd(communication.logEntry, string(plainText), true)
		if err != nil {
			return nil, errp.WithContext(err, errp.Context{"reply": string(plainText)})
		}
		jsonResult = map[string]interface{}{}
		if err := json.Unmarshal(plainText, &jsonResult); err != nil {
			return nil, errp.Wrap(err, "Failed to unmarshal reply")
		}
	}
	if err := maybeDBBErr(jsonResult); err != nil {
		return nil, err
	}
	return jsonResult, nil
}