package parser

import (
	"bufio"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

const (
	LocationTanker = "tanker"
)

func parseVector(callsign string, locations []string, scanner *bufio.Scanner) (*brevity.VectorRequest, bool) {
	request := &brevity.VectorRequest{Callsign: callsign}
	locations = append(locations, LocationTanker)

	var words []string
	for scanner.Scan() {
		word := strings.ToLower(scanner.Text())
		words = append(words, word)
	}

	for i := range words {
		for j := i; j < len(words); j++ {
			sequence := strings.Join(words[i:j+1], " ")
			for _, location := range locations {
				if isSimilar(sequence, location) {
					request.Location = location
					return request, true
				}
			}
		}
	}

	return nil, false
}
