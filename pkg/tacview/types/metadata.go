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
	duration := time.Duration(seconds*1000) * time.Millisecond
	return duration, nil
}

type ObjectUpdate struct {
	ID         uint64
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

	id, err := parseID(idStr)
	if err != nil {
		return nil, err
	}
	update.ID = id

	properties, err := parseProperties(propertiesStr)
	if err != nil {
		return nil, err
	}
	update.Properties = properties

	return update, nil
}

func parseID(idStr string) (uint64, error) {
	id, err := strconv.ParseUint(idStr, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing object ID: %w", err)
	}
	if id > math.MaxInt {
		return 0, fmt.Errorf("object ID is too large: %d", id)
	}
	return id, nil
}

func parseProperties(propertiesStr string) (map[string]string, error) {
	properties := make(map[string]string)
	if propertiesStr == "" {
		return properties, nil
	}
	if strings.Contains(propertiesStr, `\,`) {
		propertiesStr = strings.ReplaceAll(propertiesStr, `\,`, "?")
	}
	for _, prop := range strings.Split(propertiesStr, ",") {
		key, value, ok := strings.Cut(prop, "=")
		if !ok {
			return nil, fmt.Errorf("error parsing property: %s", prop)
		}
		properties[key] = strings.TrimSpace(value)
	}
	return properties, nil
}
