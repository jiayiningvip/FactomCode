// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/FactomProject/FactomCode/common"
)

// MsgMissing is used to request missing msg, ack and blocks during or after
// process list and building blocks
type MsgMissing struct {
	Height    uint32 //DBHeight for this process list
	Index     uint32 //offset in this process list
	Type      byte   //See Ack / msg types and InvTypes of blocks
	ReqNodeID string // requestor's nodeID
	Sig       common.Signature
}

// Sign is used to sign this message
func (msg *MsgMissing) Sign(priv *common.PrivateKey) error {
	bytes, err := msg.GetBinaryForSignature()
	if err != nil {
		return err
	}
	msg.Sig = priv.Sign(bytes)
	return nil
}

// GetBinaryForSignature Writes out the MsgMissing (excluding Signature) to binary.
func (msg *MsgMissing) GetBinaryForSignature() (data []byte, err error) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, msg.Height)
	binary.Write(&buf, binary.BigEndian, msg.Index)
	buf.WriteByte(msg.Type)
	buf.Write([]byte(msg.ReqNodeID))
	return buf.Bytes(), err
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgMissing) BtcDecode(r io.Reader, pver uint32) error {
	//err := readElements(r, &msg.Height, msg.ChainID, &msg.Index, &msg.Type, msg.Affirmation, &msg.SerialHash, &msg.Signature)
	newData, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("MsgMissing.BtcDecode reader is invalid")
	}

	msg.Height, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	msg.Index, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	msg.Type, newData = newData[0], newData[1:]
	len := len(newData) - 96
	msg.ReqNodeID = string(newData[:len])
	msg.Sig = common.UnmarshalBinarySignature(newData[len:])
	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgMissing) BtcEncode(w io.Writer, pver uint32) error {
	//err := writeElements(w, msg.Height, msg.ChainID, msg.Index, msg.Type, msg.Affirmation, msg.SerialHash, msg.Signature)
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, msg.Height)
	binary.Write(&buf, binary.BigEndian, msg.Index)
	buf.WriteByte(msg.Type)
	buf.Write([]byte(msg.ReqNodeID))
	data := common.MarshalBinarySignature(msg.Sig)
	buf.Write(data[:])
	w.Write(buf.Bytes())
	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgMissing) Command() string {
	return CmdMissing
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgMissing) MaxPayloadLength(pver uint32) uint32 {
	return 255
}

// NewMsgMissing returns a new bitcoin ping message that conforms to the Message
// interface.  See MsgMissing for details.
func NewMsgMissing(height uint32, index uint32, typ byte, sourceID string) *MsgMissing {
	return &MsgMissing{
		Height:    height,
		Index:     index,
		Type:      typ,
		ReqNodeID: sourceID,
	}
}

// Sha Creates a sha hash from the message binary (output of BtcEncode)
func (msg *MsgMissing) Sha() (ShaHash, error) {
	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha ShaHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))
	return sha, nil
}

// IsEomAck checks if it's a EOM ack
func (msg *MsgMissing) IsEomAck() bool {
	if END_MINUTE_1 <= msg.Type && msg.Type <= END_MINUTE_10 {
		return true
	}
	return false
}

// Equals check if two MsgMissings are the same
func (msg *MsgMissing) Equals(ack *MsgMissing) bool {
	return msg.Height == ack.Height &&
		msg.Index == ack.Index &&
		msg.Type == ack.Type &&
		msg.ReqNodeID == ack.ReqNodeID &&
		msg.Sig.Equals(ack.Sig)
}