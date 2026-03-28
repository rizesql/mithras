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

type PostID string

func NewPostID() PostID { return PostID(New("pst")) }
