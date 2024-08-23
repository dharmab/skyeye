package types

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func ParseTimeFrame(line string) (time.Duration, error) {
	if !strings.HasPrefix(line, "#") {
		return 0, fmt.Errorf("line does not contain TimeFrame: %s", line)
	}
	seconds, err := strconv.ParseFloat(strings.TrimPrefix(line, "#"), 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing duration: %w", err)
	}
	duration := time.Duration(seconds) * time.Second
	return duration, nil
}

type ObjectUpdate struct {
	ID         uint64
	IsGlobal   bool
	IsRemoval  bool
	Properties map[string]string
}

const GlobalObjectID = 0

func ParseObjectUpdate(line string) (*ObjectUpdate, error) {
	update := &ObjectUpdate{}

	if strings.HasPrefix(line, "-") {
		update.IsRemoval = true
		line = line[1:]
	}

	idStr, propertiesStr, _ := strings.Cut(line, ",")
	id, err := strconv.ParseUint(idStr, 16, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing object ID: %w", err)
	}
	if id > math.MaxInt {
		return nil, fmt.Errorf("object ID is too large: %d", id)
	}
	update.ID = id
	if id == GlobalObjectID {
		update.IsGlobal = true
	}

	update.Properties = make(map[string]string)
	if propertiesStr != "" {
		if strings.Contains(propertiesStr, `\,`) {
			propertiesStr = strings.ReplaceAll(propertiesStr, `\,`, "?")
		}
		for _, prop := range strings.Split(propertiesStr, ",") {
			key, value, ok := strings.Cut(prop, "=")
			if !ok {
				return nil, fmt.Errorf("error parsing property: %s", prop)
			}
			update.Properties[key] = strings.TrimSpace(value)
		}
	}

	return update, nil
}
