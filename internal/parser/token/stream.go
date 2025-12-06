// Package token provides a token stream implementation for parsing with backtracking support.
package token

import "strings"

// Token represents a parsed word from the input text.
type Token struct {
	Text     string // The actual text (normalized)
	Position int    // Position in the token stream
}

// Stream wraps a slice of tokens with position tracking and backtracking support.
type Stream struct {
	tokens       []Token
	pos          int
	partialToken string // Remaining text from a partially consumed token
}

// New creates a new token stream from the given normalized text.
func New(normalizedText string) *Stream {
	fields := strings.Fields(normalizedText)
	tokens := make([]Token, len(fields))
	for i, field := range fields {
		tokens[i] = Token{
			Text:     field,
			Position: i,
		}
	}
	return &Stream{tokens: tokens, pos: 0}
}

// Current returns the current token and whether it exists.
// If there's a partial token buffered, returns that instead.
func (s *Stream) Current() (Token, bool) {
	if s.partialToken != "" {
		return Token{Text: s.partialToken, Position: s.pos}, true
	}
	if s.pos >= len(s.tokens) {
		return Token{}, false
	}
	return s.tokens[s.pos], true
}

// Peek returns the token at the given offset from the current position.
// Offset 0 returns the current token, 1 returns the next, etc.
func (s *Stream) Peek(offset int) (Token, bool) {
	idx := s.pos + offset
	if idx < 0 || idx >= len(s.tokens) {
		return Token{}, false
	}
	return s.tokens[idx], true
}

// Advance moves to the next token. Returns false if at the end.
// Clears any partial token before advancing.
func (s *Stream) Advance() bool {
	s.partialToken = ""
	if s.pos >= len(s.tokens) {
		return false
	}
	s.pos++
	return s.pos < len(s.tokens)
}

// SavePosition returns the current position for later restoration.
func (s *Stream) SavePosition() int {
	return s.pos
}

// RestorePosition restores a previously saved position.
func (s *Stream) RestorePosition(pos int) {
	if pos >= 0 && pos <= len(s.tokens) {
		s.pos = pos
	}
}

// AtEnd returns true if the stream is at or past the end.
func (s *Stream) AtEnd() bool {
	return s.partialToken == "" && s.pos >= len(s.tokens)
}

// Remaining returns all tokens from the current position to the end.
func (s *Stream) Remaining() []Token {
	if s.pos >= len(s.tokens) {
		return []Token{}
	}
	return s.tokens[s.pos:]
}

// RemainingText returns the text of all remaining tokens joined by spaces.
func (s *Stream) RemainingText() string {
	remaining := s.Remaining()
	if len(remaining) == 0 {
		return ""
	}
	texts := make([]string, len(remaining))
	for i, token := range remaining {
		texts[i] = token.Text
	}
	return strings.Join(texts, " ")
}

// Skip advances past tokens that match the given predicate.
func (s *Stream) Skip(predicate func(string) bool) {
	for !s.AtEnd() {
		token, ok := s.Current()
		if !ok || !predicate(token.Text) {
			return
		}
		s.Advance()
	}
}

// Text returns the text at the current position, or empty string if at end.
func (s *Stream) Text() string {
	token, ok := s.Current()
	if !ok {
		return ""
	}
	return token.Text
}

// SetPartialToken sets the remaining text from a partially consumed token.
// This allows consuming only part of a token (e.g., "13015" -> "130" consumed, "15" remains).
func (s *Stream) SetPartialToken(remaining string) {
	s.partialToken = remaining
}

// ConsumeChars consumes n characters from the current token and sets the rest as partial.
// Returns the consumed text and whether the operation succeeded.
func (s *Stream) ConsumeChars(n int) (string, bool) {
	text := s.Text()
	if len(text) < n {
		return "", false
	}
	consumed := text[:n]
	remaining := text[n:]
	if remaining != "" {
		s.partialToken = remaining
	} else {
		s.Advance()
	}
	return consumed, true
}
