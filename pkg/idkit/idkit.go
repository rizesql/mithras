// Package idkit provides utility functions for generating prefixed NanoIDs.
package idkit

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
const length = 20

// New generates a prefixed NanoID (e.g., "req_4Fk9jP...")
func New(prefix string) string {
	id := gonanoid.MustGenerate(alphabet, length)
	return prefix + "_" + id
}

type UserID string

func NewUserID() UserID         { return UserID(New("usr")) }
func (u UserID) String() string { return string(u) }

type SessionID string

func NewSessionID() SessionID      { return SessionID(New("ses")) }
func (s SessionID) String() string { return string(s) }

type KeyID string

func NewKeyID() KeyID          { return KeyID(New("key")) }
func (k KeyID) String() string { return string(k) }

type ClientID string

func NewClientID() ClientID       { return ClientID(New("cli")) }
func (c ClientID) String() string { return string(c) }

type AuthorizationCodeID string

func NewAuthorizationCodeID() AuthorizationCodeID { return AuthorizationCodeID(New("auc")) }
func (a AuthorizationCodeID) String() string      { return string(a) }

type PasswordResetID string

func NewPasswordResetID() PasswordResetID { return PasswordResetID(New("rst")) }
func (p PasswordResetID) String() string  { return string(p) }
